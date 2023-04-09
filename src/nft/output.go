package nft

import (
	"fmt"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/wire"
)

type OutputData struct {
	Output wire.OutPoint
	Amount btcutil.Amount
}

func OutputRun(opt *Options) error {
	index, err := Open(opt)
	if err != nil {
		return err
	}
	Update(index)
	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

	var res []OutputData
	for output, amount := range unspentOutput {
		res = append(res, OutputData{
			Output: output,
			Amount: amount,
		})
	}

	fmt.Print(res)
	return nil
}
