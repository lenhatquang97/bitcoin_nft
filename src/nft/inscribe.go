package nft

import (
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/m25lab/bitcoin_nft/src"
)

type Output struct {
	Commit        string
	InscriptionID src.InscriptionId
	Reveal        string
	Fee           int64
}

type Inscribe struct {
	SatPoint      src.SatPoint
	FeeRate       float64
	CommitFeeRate float64
	File          string
	NoBackup      bool
	NoLimit       bool
	DryRun        bool
	Destination   btcutil.Address
}

func Run(inscribe *Inscribe, opt *Options) error {
	inscription, err := NftFromFile(inscribe.File)
	if err != nil {
		return err
	}

	index, err := Open(opt)
	if err != nil {
		return err
	}

	client, err := GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		return err
	}

	utxos, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

	inscriptions, err := GetInscription(index)
	if err != nil {
		return err
	}

	firstAddress, err := client.GetRawChangeAddressType("", "bech32m")
	if err != nil {
		return err
	}

	secondAddress, err := client.GetRawChangeAddressType("", "bech32m")
	if err != nil {
		return err
	}

	var commitTxChange []btcutil.Address
	commitTxChange = append(commitTxChange, firstAddress)
	commitTxChange = append(commitTxChange, secondAddress)

	var revealTxDestination btcutil.Address
	if inscribe.Destination != nil {
		revealTxDestination = inscribe.Destination
	} else {
		// handle error
		revealTxDestination, _ = client.GetRawChangeAddressType("", "bech32m")
	}

	commitFeeRate := inscribe.CommitFeeRate
	if commitFeeRate < 0 {
		commitFeeRate = inscribe.FeeRate
	}
	unsignedCommitTx, revealTx, err := CreateInscriptionTransaction(&inscribe.SatPoint, inscription, inscriptions, &chaincfg.TestNet3Params, utxos, commitTxChange, revealTxDestination, commitFeeRate, inscribe.FeeRate, inscribe.NoLimit)
	if err != nil {
		return err
	}

	utxos[revealTx.TxIn[0].PreviousOutPoint] = btcutil.Amount(unsignedCommitTx.TxOut[0].Value)

	fees := CalculateFee(unsignedCommitTx, utxos) + CalculateFee(revealTx, utxos)
	if inscribe.DryRun {
		// log result output
		output := Output{
			Commit:        unsignedCommitTx.TxHash().String(),
			Reveal:        revealTx.TxHash().String(),
			InscriptionID: src.InscriptionId{},
			Fee:           int64(fees),
		}

		log.Println(output)
	} else {
		signRawCommitTx, _, err := client.SignRawTransactionWithWallet(unsignedCommitTx)
		if err != nil {
			return err
		}

		commit, err := client.SendRawTransaction(signRawCommitTx, false)
		if err != nil {
			return err
		}

		reveal, err := client.SendRawTransaction(revealTx, false)
		if err != nil {
			return err
		}

		output := Output{
			Commit:        commit.String(),
			Reveal:        reveal.String(),
			InscriptionID: src.InscriptionId{},
			Fee:           int64(fees),
		}

		log.Println(output)
	}

	return nil
}
