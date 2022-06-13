package main

import (
	"geth-block-event-listener/constants"
	"geth-block-event-listener/ethereum"
	"os"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	client := ethereum.Connect()
	wg := new(sync.WaitGroup)
	wg.Add(1)
	if len(constants.ERC20ContractAddressesToMonitor) > 0 {
		wg.Add(1)
		go ethereum.ListenerForERC20Transfers(client, constants.ERC20ContractAddressesToMonitor, wg)
	}
	go ethereum.ListenerForEthereumTransfers(client, wg)
	wg.Wait()
}
