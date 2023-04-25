package layer1

import (
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
)

func GetOutputIndex(txid *chainhash.Hash, addr string, client *rpcclient.Client) int {
	tx, _ := client.GetRawTransaction(txid)

	outputIndex := -1
	for i, txOut := range tx.MsgTx().TxOut {
		_, addrs, _, _ := txscript.ExtractPkScriptAddrs(
			txOut.PkScript, &chaincfg.TestNet3Params,
		)

		if addrs[0].String() == addr {
			outputIndex = i
		}
	}
	return outputIndex
}
