package wallet

import (
	"encoding/json"
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

// no use
func GetBalance(opt *nft.Options) (*btcutil.Amount, error) {
	client, err := nft.GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	result, err := client.GetBalance("")
	return &result, err
}

// no use
func GetAddress(opt *nft.Options) (*btcutil.Address, error) {
	client, err := nft.GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		return nil, err
	}
	address, err := client.GetNewAddress("")
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// no use
func PrintUnspentOutputs(opt *nft.Options) {
	client, err := nft.GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		fmt.Println(err)
		return
	}
	results, err := client.ListUnspent()
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, result := range results {
		marshaled, err := json.MarshalIndent(result, "", " ")
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(marshaled))
	}

}
