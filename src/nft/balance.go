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

	var satPoints []*wire.OutPoint
	for satPoint := range inscriptiOutput {
		satPoints = append(satPoints, &satPoint.OutPoint)
	}

	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

	var balance btcutil.Amount
	for _, amount := range unspentOutput {
		balance += amount
	}

	fmt.Println("Balance: {}", balance.ToBTC())
	return nil
}
