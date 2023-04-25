package main

import (
	"fmt"

	"github.com/m25lab/bitcoin_nft/src/wallet"
)

func main() {
	output, err := wallet.RetrieveTransaction()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(output)
	}
}
