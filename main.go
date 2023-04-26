package main

import (
	"fmt"

	"github.com/m25lab/bitcoin_nft/src/nft"
)

func main() {
	err := nft.Run(&nft.Inscribe{
		SatPoint:      nil,
		FeeRate:       1.1,
		CommitFeeRate: 1,
		File:          "./taro_note.txt",
		Destination:   "bcrt1qjrdns4f5zwkv29ln86plqzs092yd5fg6nsz8re",
		DryRun:        false,
	}, &nft.Options{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successful")
	}
}
