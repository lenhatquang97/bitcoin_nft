package nft

import "fmt"

type ReceiveOutput struct {
	Address string
}

/*
* Have reviewed in 10/4/2023
 */

// no use
func ReceiveRun(opt *Options) error {
	client, err := GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		return err
	}

	address, err := client.GetNewAddressType("", "bech32m")
	if err != nil {
		return err
	}

	res := &ReceiveOutput{
		Address: address.String(),
	}

	fmt.Print(res)

	return nil
}
