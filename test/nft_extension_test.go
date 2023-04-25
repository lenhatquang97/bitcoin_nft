package test

import (
	"fmt"
	"os"
	"testing"

	"github.com/btcsuite/btcd/txscript"
	"github.com/m25lab/bitcoin_nft/src/inscript"
	"github.com/m25lab/bitcoin_nft/src/model"
)

func TestWithSimplePDF(t *testing.T) {
	inscription, _ := inscript.NftFromFile("./Consensus4.pdf")
	script := inscript.NftRevealScript(inscription, *txscript.NewScriptBuilder())
	result := inscript.ParseScriptToInscription(script)
	err := os.WriteFile("./final.pdf", result.Body, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result.ContentType)
}

func TestWithSimpleText(t *testing.T) {
	inscription := model.Inscription{
		ContentType: "text/plain",
		Body:        []byte("Hello World"),
	}
	script := inscript.NftRevealScript(&inscription, *txscript.NewScriptBuilder())
	result := inscript.ParseScriptToInscription(script)

	actualString := string(result.Body)
	expectedString := "Hello World"
	if actualString != expectedString {
		t.Errorf("Wrong! Expected string is %s, but actual string is %s", expectedString, actualString)
	}
}
