package nft

import (
	"fmt"
	"github.com/m25lab/bitcoin_nft/src/mnemonic"
)

type CreateData struct {
	Mnemonic   string
	Passphrase string
}

type CreateInto struct {
	PassPhrase string
}

func CreateRun(input *CreateInto, opt *Options) error {
	// gen mnemonic
	keyManager, err := mnemonic.NewKeyManager(128, input.PassPhrase, "")
	if err != nil {
		return err
	}

	// to seed
	seed := keyManager.GetSeed()
	output := &CreateData{
		Mnemonic:   keyManager.GetMnemonic(),
		Passphrase: keyManager.GetPassphrase(),
	}

	fmt.Print("Create result: ", output)

	// initialize wallet
	return InitializeWallet(opt, seed)
}
