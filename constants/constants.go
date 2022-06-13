package constants

import (
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	DAI  = "0x6B175474E89094C44Da98b954EedeAC495271d0F"
	ILV  = "0x767fe9edc9e0df98e07454847909b5e959d7ca0e"
	VRA  = "0xf411903cbc70a74d22900a5de66a2dda66507255"
	USDT = "0xdac17f958d2ee523a2206206994597c13d831ec7"
)

var (
	LogTransferSig     = []byte("Transfer(address,address,uint256)")
	LogTransferSigHash = crypto.Keccak256Hash(LogTransferSig)

	ERC20ContractAddressesToMonitor = []string{DAI, ILV, VRA, USDT}
)
