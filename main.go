package main

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/txscript"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

func main() {
	inscription, _ := nft.NftFromFile("./mrt.png")
	script := nft.NftRevealScript(inscription, *txscript.NewScriptBuilder())
	result := nft.ParseScriptToInscription(script)
	err := os.WriteFile("./final.png", result.Body, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result.ContentType)
}
