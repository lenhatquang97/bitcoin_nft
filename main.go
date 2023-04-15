package main

import (
	"fmt"

	"github.com/m25lab/bitcoin_nft/src/enum"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

func main() {
	option := nft.Options{
		ChainArgument:  enum.Chain.Testnet,
		BitcoinDataDir: "./bitcoindatadir/",
		DataDir:        "./datadir/",
		RpcUrl:         "localhost:8334",
		IndexSats:      true,
	}
	result, err := nft.Open(&option)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result.FirstInscriptionHeight)
	}
}
