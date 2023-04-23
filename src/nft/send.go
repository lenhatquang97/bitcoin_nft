package nft

import (
	"context"
	"errors"
	"fmt"
	"github.com/lightningnetwork/lnd/lnrpc"
	"log"
	"time"

	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/enum"
)

type SendData struct {
	Address          string
	OutGoingType     enum.OutGoingTypeValue
	OutGoingTypeData interface{}
	FeeRate          float64
}

type SendDataOutput struct {
	TxID string
}

func SendRun(opt *Options, data *SendData) error {
	//if !data.Address.IsForNet(&chaincfg.TestNet3Params) {
	//	return errors.New("Address is not valid")
	//}

	// Check testnet address

	index, err := Open(opt)
	if err != nil {
		return err
	}
	Update(index)
	client, err := GetBitcoinRPCClientForWalletCommand(opt, false)
	if err != nil {
		return err
	}

	unspentOutput, err := GetUnspentOutput(index)
	if err != nil {
		return err
	}

	inscriptions, err := GetInscription(index)
	if err != nil {
		return err
	}

	outGoing := new(src.SatPoint)
	lndConn, err := GetLndGrpcSetup()
	if err != nil {
		return err
	}
	defer lndConn.Close()

	lncli := lnrpc.NewLightningClient(lndConn)

	switch data.OutGoingType {
	case enum.OutGoingType.Satpoint:
		satpoint := data.OutGoingTypeData.(*src.SatPoint)
		if satpoint == nil {
			return errors.New("Satpoint data is invalid")
		}

		for inscriptionSatpoint := range inscriptions {
			if satpoint.Serialize() == inscriptionSatpoint {
				s := "inscription must be spent by inscription id"
				fmt.Println(s)
				return errors.New(s)
			}
		}
		outGoing = satpoint
	case enum.OutGoingType.InscriptionId:
		inscriptionId := data.OutGoingTypeData.(*src.InscriptionId)
		if inscriptionId == nil {
			return errors.New("inscription id data is invalid")
		}

		outGoing, err = GetInscriptionSatPointById(index, inscriptionId)
		if err != nil {
			log.Printf("inscription %v not found", inscriptionId)
			return err
		}
	case enum.OutGoingType.Amount:
		allInscriptionOutput := make(map[string]string)
		for inscriptionOuput := range inscriptions {
			satpointDeserialize, err := src.DeserializeSatPoint(inscriptionOuput)
			if err != nil {
				return err
			}
			allInscriptionOutput[satpointDeserialize.OutPoint.Serialize()] = ""
		}

		var walletInscriptionOutput []*Outpoint
		for utxo := range unspentOutput {
			_, ok := allInscriptionOutput[utxo]
			if ok {
				deserializeOutpoint, err := DeserializeOutpoint(utxo)
				if err != nil {
					return err
				}
				walletInscriptionOutput = append(walletInscriptionOutput, deserializeOutpoint)
			}
		}

		amount := data.OutGoingTypeData.(int64)
		//err = client.LockUnspent(false, walletInscriptionOutput)
		//if err != nil {
		//	return err
		//}

		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		res, err := lncli.SendCoins(ctx, &lnrpc.SendCoinsRequest{
			Addr:   data.Address,
			Amount: amount,
		})
		if err != nil {
			return err
		}

		fmt.Println("Chain hash: ", res)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	newAddressReq := lnrpc.NewAddressRequest{Type: lnrpc.AddressType_TAPROOT_PUBKEY}

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
	unsignedCommitTx, err := BuildTransactionWithPostage(*outGoing, inscriptions, unspentOutput, data.Address, commitTxChange, data.FeeRate)
	if err != nil {
		return err
	}

	signedTx, _, err := client.SignRawTransactionWithWallet(unsignedCommitTx)
	if err != nil {
		return err
	}

	res, err := client.SendRawTransaction(signedTx, false)
	if err != nil {
		return err
	}
	fmt.Println("Tx hash: ", res)
	return nil
}
