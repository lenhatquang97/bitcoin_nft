package main

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"

	"github.com/m25lab/bitcoin_nft/src/nft"
)

func main() {
	addr, err := btcutil.DecodeAddress("SeTCfjeSQYevShUDEqo59GH1V5kqnP4dg5", &chaincfg.SimNetParams)
	if err != nil {
		fmt.Println(err)
	}

	err = nft.Run(&nft.Inscribe{
		SatPoint:      nil,
		FeeRate:       1.1,
		CommitFeeRate: 1,
		File:          "./taro_note.txt",
		Destination:   addr,
		DryRun:        false,
	}, &nft.Options{})
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Successful")
	}
}
