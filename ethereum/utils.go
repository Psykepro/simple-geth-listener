package ethereum

import (
	"context"
	"fmt"
	token "geth-block-event-listener/contracts"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/rs/zerolog/log"
)

var (
	ERC20TokensByAddress = make(map[common.Address]*token.Token)
)

func getTokenInstance(address common.Address, ethClient *ethclient.Client) (*token.Token, bool) {
	instance, ok := ERC20TokensByAddress[address]
	if !ok {
		var err error
		instance, err = token.NewToken(address, ethClient)
		if err != nil {
			log.Error().Msg(fmt.Sprintf("failed to initialize token instance for contract address: %s", address.String()))
			return nil, false
		}
		ERC20TokensByAddress[address] = instance

	}

	return instance, true
}

func logTokenSymbol(client *ethclient.Client, contractAddress common.Address) {
	tokenInstance, ok := getTokenInstance(contractAddress, client)
	if ok {
		symbol, err := tokenInstance.Symbol(new(bind.CallOpts))
		if err == nil {
			log.Info().Str("Symbol", symbol).Msg("")
		}
	}
}

func subscribeForERC20Transfers(client *ethclient.Client, addresses []common.Address) (chan types.Log, ethereum.Subscription) {
	query := ethereum.FilterQuery{Addresses: addresses}
	logs := make(chan types.Log)
	log.Info().Msg("trying to subscribe for ERC20 transfers ...")
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to subscribe for ERC20 transfers. Err: %s", err.Error()))
	}
	log.Info().Msg("successfully subscribed to listen for ERC20 transfers")

	return logs, sub
}

func subscribeForAllTransactionHeaders(client *ethclient.Client) (chan *types.Header, ethereum.Subscription) {
	headers := make(chan *types.Header)
	log.Info().Msg("trying to subscribe for headers ...")
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("failed to subscribe for headers. Err: %s", err.Error()))
	}
	log.Info().Msg("successfully subscribed to listen for headers")
	return headers, sub
}

func logBasicBlockInfo(block *types.Block, headerPrefix string) {
	msg := fmt.Sprintf("# %sBlock with hash: %s #", headerPrefix, block.Hash().Hex())
	delimiterLine := strings.Repeat("#", len(msg))
	log.Info().Msg(delimiterLine)
	log.Info().Msg(msg)
	log.Info().Msg(delimiterLine)
	log.Info().Uint64("Number", block.Number().Uint64()).Msg("")
	log.Info().Uint64("Time", block.Time()).Msg("")
	log.Info().Uint64("Nonce", block.Nonce()).Msg("")
	log.Info().Int("Transactions Count", len(block.Transactions())).Msg("")

	totalEtherInWei := big.NewInt(0)
	for _, transaction := range block.Transactions() {
		totalEtherInWei = big.NewInt(0).Add(totalEtherInWei, transaction.Value())
	}
	log.Info().Str("Ether in wei", totalEtherInWei.String()).Msg("")
	totalEtherInWeiFloat, _ := strconv.ParseFloat(totalEtherInWei.String(), 128)
	totalEther := new(big.Float).Quo(big.NewFloat(totalEtherInWeiFloat), big.NewFloat(params.Ether)).String()
	log.Info().Str("Ether", totalEther).Msg("")
}

func Connect() *ethclient.Client {
	log.Info().Msg("ethclient trying to connect ...")
	client, err := ethclient.Dial("ws://localhost:3334")
	if err != nil {
		log.Fatal().Msg(fmt.Sprintf("ethclient failed to connect. Err: %s", err.Error()))
	}
	log.Info().Msg("ethclient successfully connected")

	return client
}
