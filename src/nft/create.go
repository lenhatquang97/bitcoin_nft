package nft

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcwallet/wallet"
	"github.com/btcsuite/btcwallet/walletdb"
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
	privPass := []byte("password")
	pubPass := []byte(wallet.InsecurePubPassphrase)

	basePath := btcutil.AppDataDir(opt.DataDir, false)
	dbPath := filepath.Join(basePath, wallet.WalletDBName)
	fmt.Println("Creating the wallet...")

	db, err := walletdb.Create("bbolt", dbPath, true)
	if err != nil {
		return err
	}
	defer db.Close()

	err = wallet.Create(db, pubPass, privPass, nil, &chaincfg.TestNet3Params, time.Now())
	if err != nil {
		return err
	}

	fmt.Println("The wallet has been created successfully.")
	return nil
}
