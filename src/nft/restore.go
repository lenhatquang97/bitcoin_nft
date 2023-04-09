package nft

import "github.com/m25lab/bitcoin_nft/src/mnemonic"

type Restore struct {
	Mnemonic   string
	Passphrase string
}

func RestoreRun(opt *Options, restore *Restore) error {
	// initialize wallet by restore
	keyManager, err := mnemonic.NewKeyManager(128, restore.Passphrase, "")
	if err != nil {
		return err
	}

	seed := keyManager.GetSeed()
	return InitializeWallet(opt, seed)
}
