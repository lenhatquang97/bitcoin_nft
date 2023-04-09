package nft

import "fmt"

type ReceiveOutput struct {
	Address string
}

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
