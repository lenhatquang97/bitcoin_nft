package nft

import (
	"github.com/m25lab/bitcoin_nft/src/mnemonic"
)

type CreateData struct {
	Mnemonic   string
	Passphrase string
}

/*
* Have reviewed in 9/4/2023
 */

func GenMnemonicAndSeed(passPhrase string) ([]byte, *CreateData, error) {
	// gen mnemonic
	keyManager, err := mnemonic.NewKeyManager(128, passPhrase, "")
	if err != nil {
		return nil, nil, err
	}

	// to seed
	seed := keyManager.GetSeed()
	output := &CreateData{
		Mnemonic:   keyManager.GetMnemonic(),
		Passphrase: keyManager.GetPassphrase(),
	}

	return seed, output, nil
}

func CreateRun(passPhrase string, opt *Options) error {
	return nil
}
