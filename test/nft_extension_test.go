package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/btcsuite/btcd/txscript"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

func TestWithSimplePDF(t *testing.T) {
	inscription, _ := nft.NftFromFile("./Consensus4.pdf")
	script := nft.NftRevealScript(inscription, *txscript.NewScriptBuilder())
	result := nft.ParseScriptToInscription(script)
	err := os.WriteFile("./final.pdf", result.Body, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result.ContentType)
}

func TestWithSimpleText(t *testing.T) {
	inscription := nft.Inscription{
		ContentType: "text/plain",
		Body:        []byte("Hello World"),
	}
	script := nft.NftRevealScript(&inscription, *txscript.NewScriptBuilder())
	result := nft.ParseScriptToInscription(script)

	actualString := string(result.Body)
	expectedString := "Hello World"
	if actualString != expectedString {
		t.Errorf("Wrong! Expected string is %s, but actual string is %s", expectedString, actualString)
	}
}
