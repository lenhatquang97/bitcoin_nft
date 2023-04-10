package nft

import (
	"fmt"

	"github.com/btcsuite/btcd/rpcclient"
)

func InitializeWallet(opt *Options, seed []byte) error {
	client, err := GetBitcoinRPCClientForWalletCommand(opt, true)
	if err != nil {
		return err
	}

	res, err := client.CreateWallet(opt.Wallet, rpcclient.WithCreateWalletBlank())
	if err != nil {
		return err
	}

	fmt.Print(res)
	// derive and import descriptor

	return nil
}
