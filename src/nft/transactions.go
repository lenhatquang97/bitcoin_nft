package nft

import "fmt"

type TransactionOutput struct {
	TxID          string
	Confirmations int64
}

/*
* No need to review
 */

func TransactionRun(opt *Options) error {
	client, err := GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		return err
	}
	transactions, err := client.ListTransactions("*")
	if err != nil {
		return err
	}

	var res []TransactionOutput
	for _, tx := range transactions {
		res = append(res, TransactionOutput{
			TxID:          tx.TxID,
			Confirmations: tx.Confirmations,
		})
	}

	fmt.Print(res)
	return nil
}
