package nft

import (
	"context"
	"log"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/lightningnetwork/lnd/lnrpc"
)

type Output struct {
	Commit        string
	InscriptionID InscriptionId
	Reveal        string
	Fee           int64
}

type Inscribe struct {
	SatPoint      SatPoint
	FeeRate       float64
	CommitFeeRate float64
	File          string
	NoBackup      bool
	NoLimit       bool
	DryRun        bool
	Destination   string
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

	lndConn, err := GetLndGrpcSetup()
	if err != nil {
		return err
	}
	defer lndConn.Close()

	lncli := lnrpc.NewLightningClient(lndConn)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	newAddressReq := lnrpc.NewAddressRequest{Type: lnrpc.AddressType_TAPROOT_PUBKEY}

	utxos, err := GetUnspentOutput()
	if err != nil {
		return err
	}

	inscriptions, err := GetInscription(index)
	if err != nil {
		return err
	}

	firstAddress, err := lncli.NewAddress(ctx, &newAddressReq) // bech32m
	if err != nil {
		return err
	}

	secondAddress, err := lncli.NewAddress(ctx, &newAddressReq) // bech32m
	if err != nil {
		return err
	}

	var commitTxChange []string
	commitTxChange = append(commitTxChange, firstAddress.Address)
	commitTxChange = append(commitTxChange, secondAddress.Address)

	revealTxDestination := ""
	if inscribe.Destination != "" {
		revealTxDestination = inscribe.Destination
	} else {
		// handle error
		revealAddRes, err := lncli.NewAddress(ctx, &newAddressReq)
		if err != nil {
			return err
		}
		revealTxDestination = revealAddRes.Address
	}

	commitFeeRate := inscribe.CommitFeeRate
	if commitFeeRate < 0 {
		commitFeeRate = inscribe.FeeRate
	}
	unsignedCommitTx, revealTx, err := CreateInscriptionTransaction(&inscribe.SatPoint, inscription, inscriptions, &chaincfg.TestNet3Params, utxos, commitTxChange, revealTxDestination, commitFeeRate, inscribe.FeeRate, inscribe.NoLimit)
	if err != nil {
		return err
	}

	outpointConverted := ConvertToOutpoint(&revealTx.TxIn[0].PreviousOutPoint)
	utxos[outpointConverted.Serialize()] = unsignedCommitTx.TxOut[0].Value

	fees := CalculateFee(unsignedCommitTx, utxos) + CalculateFee(revealTx, utxos)
	if inscribe.DryRun {
		// log result output
		output := Output{
			Commit:        unsignedCommitTx.TxHash().String(),
			Reveal:        revealTx.TxHash().String(),
			InscriptionID: InscriptionId{},
			Fee:           int64(fees),
		}

		log.Println(output)
	} else {
		unsignedCommitTxHash := unsignedCommitTx.TxHash()

		signRawCommitTx, err := lncli.SignMessage(ctx, &lnrpc.SignMessageRequest{
			Msg: unsignedCommitTxHash.CloneBytes(),
		})
		if err != nil {
			return err
		}

		commit, err := lncli.SendCustomMessage(ctx, &lnrpc.SendCustomMessageRequest{
			Data: []byte(signRawCommitTx.String()),
			Peer: []byte(revealTxDestination),
		})

		if err != nil {
			return err
		}

		revealTxHash := revealTx.TxHash()
		reveal, err := lncli.SendCustomMessage(ctx, &lnrpc.SendCustomMessageRequest{
			Data: revealTxHash.CloneBytes(),
			Peer: []byte(revealTxDestination),
		})

		if err != nil {
			return err
		}

		output := Output{
			Commit:        commit.String(),
			Reveal:        reveal.String(),
			InscriptionID: InscriptionId{},
			Fee:           int64(fees),
		}

		log.Println(output)
	}

	return nil
}
