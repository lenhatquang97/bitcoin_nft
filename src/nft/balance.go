package nft

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

func BalanceRun(opt *Options) error {
	index, err := Open(opt)
	if err != nil {
		return err
	}

	inscriptiOutput, err := GetInscription(index)
	if err != nil {
		return err
	}

	satPoints := make(map[wire.OutPoint]string)
	for satPoint := range inscriptiOutput {
		satPoints[satPoint.OutPoint] = ""
	}

	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

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
