package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/m25lab/bitcoin_nft/src/nft"
)

func TestValidMnemonic(t *testing.T) {
	seed, output, err := nft.GenMnemonicAndSeed("")
	if err != nil {
		t.Error(err)
	} else {
		fmt.Println(seed)
		mnemonicWords := strings.Split(output.Mnemonic, " ")
		if len(mnemonicWords) != 12 {
			t.Errorf("Expected mnemonic words must be 12 words but actual words are %d", len(mnemonicWords))
		}
	}
}
