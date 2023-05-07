package nft

import (
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
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
	SatPoint      *src.SatPoint
	FeeRate       float64
	CommitFeeRate float64
	File          string
	NoBackup      bool
	NoLimit       bool
	DryRun        bool
	Destination   btcutil.Address
}

const FirstSeed = "d94155d877b8150f6215ad5bc6917989fd88888c045a21791fed17e0ae916bec"
const FirstMiningAddress = "SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi"

func GetPayToAddrScript(address string) []byte {
	rcvAddress, _ := btcutil.DecodeAddress(address, &chaincfg.SimNetParams)
	rcvScript, _ := txscript.PayToAddrScript(rcvAddress)
	return rcvScript
}

func GetPrivateKey(privKey string) (*btcec.PrivateKey, *btcec.PublicKey, error) {
	privByte, err := hex.DecodeString(privKey)

	if err != nil {
		return nil, nil, err
	}

	priv, pubKey := btcec.PrivKeyFromBytes(privByte) //secp256k1
	return priv, pubKey, nil
}

func SignTx(redeemTx *wire.MsgTx, subscript []byte, privKey *secp256k1.PrivateKey) {
	sig, err := txscript.SignatureScript(redeemTx, 0, subscript, txscript.SigHashAll, privKey, false)
	if err != nil {
		log.Fatalf("could not generate signature: %v", err)
	}
	redeemTx.TxIn[0].SignatureScript = sig
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

	firstAddress, err := btcutil.DecodeAddress("SeTCfjeSQYevShUDEqo59GH1V5kqnP4dg5", &chaincfg.SimNetParams)
	if err != nil {
		return err
	}

	secondAddress, err := btcutil.DecodeAddress("SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi", &chaincfg.SimNetParams)
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
		revealTxDestination, _ = btcutil.DecodeAddress("SZnK16oMnqQt8Q1qLvrTpYLpkpkFG9eVRi", &chaincfg.SimNetParams)
	}

	commitFeeRate := inscribe.CommitFeeRate
	if commitFeeRate < 0 {
		commitFeeRate = inscribe.FeeRate
	}
	unsignedCommitTx, revealTx, err := CreateInscriptionTransaction(inscribe.SatPoint, inscription, inscriptions, &chaincfg.SimNetParams, utxos, commitTxChange, revealTxDestination, commitFeeRate, inscribe.FeeRate, inscribe.NoLimit)
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
		privKey, _, _ := GetPrivateKey(FirstSeed)
		SignTx(unsignedCommitTx, GetPayToAddrScript(FirstMiningAddress), privKey)
		signRawCommitTx := unsignedCommitTx
		signRawCommitTx, _, err := client.SignRawTransaction(unsignedCommitTx)
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
