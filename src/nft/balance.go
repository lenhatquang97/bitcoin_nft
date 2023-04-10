package nft

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

/*
* Have reviewed in 9/4/2023
* Balance:
* Step 1: Connect node
* Step 2: Get all stored NFT (inscriptions go with satpoint)
* Step 3: Balance = Total balance - stored NFT
* Needs integration test
 */
func BalanceRun(opt *Options) error {
	index, err := Open(opt)
	if err != nil {
		return err
	}

	inscriptionOutput, err := GetInscription(index)
	if err != nil {
		return err
	}

	satPoints := make(map[wire.OutPoint]string)
	for satPoint := range inscriptionOutput {
		satPoints[satPoint.OutPoint] = ""
	}

	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

	//Balances = total balance - NFT (NFT defines with satPoints)
	var balance btcutil.Amount
	for outpoint, amount := range unspentOutput {
		_, ok := satPoints[outpoint]
		if !ok {
			balance += amount
		}
	}

	fmt.Println("Balance: {}", balance.ToBTC())
	return nil
}
