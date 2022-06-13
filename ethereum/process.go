package ethereum

import (
	"context"
	"fmt"
	"geth-block-event-listener/constants"
	token "geth-block-event-listener/contracts"
	"geth-block-event-listener/models"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

func ListenerForEthereumTransfers(client *ethclient.Client, wg *sync.WaitGroup) {
	defer wg.Done()
	headers, sub := subscribeForAllTransactionHeaders(client)
	for {
		select {
		case err := <-sub.Err():
			log.Fatal().Err(err)
		case header := <-headers:
			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Error().Msg(fmt.Sprintf("failed to get full block. Err: %s", err))
				continue
			}
			logBasicBlockInfo(block, "")
		}
	}
}

func ListenerForERC20Transfers(client *ethclient.Client, contractAddresses []string, wg *sync.WaitGroup) {
	defer wg.Done()
	contractAbi, err := abi.JSON(strings.NewReader(token.TokenMetaData.ABI))
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to initialize ERC20 contract abi. Err: %s", err))
		return
	}

	addresses := make([]common.Address, len(contractAddresses))
	for i, addressAsStr := range contractAddresses {
		addresses[i] = common.HexToAddress(addressAsStr)
	}

	logs, sub := subscribeForERC20Transfers(client, addresses)
	for {
		select {
		case err := <-sub.Err():
			log.Fatal().Err(err)
		case vLog := <-logs:
			block, err := client.BlockByHash(context.Background(), vLog.BlockHash)
			if err != nil {
				log.Error().Msg(fmt.Sprintf("failed to get full block. Err: %s", err))
				continue
			}
			logBasicBlockInfo(block, "ERC-20 Transfer ")

			switch vLog.Topics[0].Hex() {
			case constants.LogTransferSigHash.Hex():
				var transferEvent models.LogTransfer
				err := contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
				if err != nil {
					log.Error().Msg(fmt.Sprintf("failed to unpack contract abi to LogTransfer model. Err: %s", err))
					continue
				}
				transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
				transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())
				log.Info().Str("From", transferEvent.From.Hex()).Msg("")
				log.Info().Str("To", transferEvent.To.Hex()).Msg("")
				log.Info().Str("Value (Tokens Amount)", transferEvent.Value.String()).Msg("")
				logTokenSymbol(client, vLog.Address)
			}
		}
	}
}
