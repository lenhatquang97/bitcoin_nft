package main

import (
	"fmt"
	"log"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/m25lab/bitcoin_nft/src/enum"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

func main() {
	option := nft.Options{
		ChainArgument:  enum.Chain.Testnet,
		BitcoinDataDir: "bitcoindatadir",
		DataDir:        "datadir",
		RpcUrl:         "localhost:18332",
		IndexSats:      true,
		Wallet:         "mywalleter",
	}
	index, err := nft.Open(&option)
	if err != nil {
		log.Fatal(err)
	}

	res, err := index.Client.CreateWallet("", rpcclient.WithCreateWalletBlank())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)
}
