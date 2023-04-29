package nft

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/database"
	"github.com/btcsuite/btcd/mempool"
	"github.com/m25lab/bitcoin_nft/src/inscript"
	"github.com/m25lab/bitcoin_nft/src/model"

	//"github.com/btcsuite/btcd/btcutil/schnorr/musig2"

	_ "github.com/btcsuite/btcd/database/ffldb"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

/*
* Have reviewed in 10/4/2023
* Remove ContentLength, BackupRecoveryKey
*
 */

// Build Reveal Transaction: Schnorr Signature is Empty
func BuildRevealTransaction(ctrlBlock *txscript.ControlBlock, feeRate float64, input *wire.OutPoint, output *wire.TxOut, script []byte) (*wire.MsgTx, btcutil.Amount) {
	emptyScript, _ := txscript.NewScriptBuilder().Script()
	emptyWitness := wire.TxWitness{}
	revealTx := wire.MsgTx{
		Version:  1,
		LockTime: 0,
		TxIn: []*wire.TxIn{
			{
				PreviousOutPoint: *input,
				SignatureScript:  emptyScript,
				Witness:          emptyWitness,
				Sequence:         0,
			},
		},
		TxOut: []*wire.TxOut{output},
	}

	copyTx := revealTx.Copy()
	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, make([]byte, schnorr.SignatureSize))
	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, script)
	ctrlBlockByte, err := ctrlBlock.ToBytes()
	if err != nil {
		fmt.Println(err)
	}
	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, ctrlBlockByte)

	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(copyTx))
	actualFee := Fee(feeRate, float64(txSize))
	return &revealTx, actualFee
}

func ToWitness(builder *txscript.ScriptBuilder) (wire.TxWitness, error) {
	script, err := builder.Script()
	if err != nil {
		return nil, err
	}
	witness := wire.TxWitness(make([][]byte, 2))
	witness[0] = script
	witness[1] = make([]byte, 0)
	return witness, nil
}

func CalculateFee(tx *wire.MsgTx, utxos map[string]int64) btcutil.Amount {
	var sumTxIn int64
	for _, v := range tx.TxIn {
		outpointConverted := model.ConvertToOutpoint(&v.PreviousOutPoint)
		sumTxIn += utxos[outpointConverted.Serialize()]
	}

	var sumTxOut int64
	for _, v := range tx.TxOut {
		sumTxOut += v.Value
	}

	return btcutil.Amount(sumTxIn - sumTxOut)
}

func CreateInscriptionTransaction(satpoint *SatPoint,
	inscription *model.Inscription,
	inscriptions map[string]InscriptionId,
	network *chaincfg.Params,
	utxos map[string]int64,
	change []string,
	destination string,
	commitFeeRate float64,
	revealFeeRate float64,
	noLimit bool,
) (*wire.MsgTx, *wire.MsgTx, error) {
	var satP *SatPoint
	if satpoint != nil {
		satP = satpoint
	} else {
		inscribeUtxos := make(map[string]string) // find about set in golang
		for inscrp := range inscriptions {
			s, err := DeserializeSatPoint(inscrp)
			if err != nil {
				return nil, nil, err
			}

			inscribeUtxos[s.OutPoint.Serialize()] = ""
		}

		for outpoint := range utxos {
			_, ok := inscribeUtxos[outpoint]
			if !ok {
				outpointConverted, err := model.DeserializeOutpoint(outpoint)
				if err != nil {
					return nil, nil, err
				}

				satP = &SatPoint{
					OutPoint: *outpointConverted,
					OffSet:   0,
				}
				break
			}
		}
	}

	satpointSerialize := satP.Serialize()
	for inscribedSatpoint, inscriptionId := range inscriptions {
		if inscribedSatpoint == satpointSerialize {
			return nil, nil, fmt.Errorf("sat at %v sat poiont already inscribed", satpoint)
		}

		satpointDeserialize, err := DeserializeSatPoint(inscribedSatpoint)
		if err != nil {
			return nil, nil, err
		}

		if satpointDeserialize.OutPoint.Serialize() == satpoint.OutPoint.Serialize() {
			return nil, nil, fmt.Errorf("utxo already inscribed %v on sat %v", inscriptionId, inscribedSatpoint)
		}
	}

	privKey, _ := btcec.NewPrivateKey()
	pubKey := privKey.PubKey()
	//revealScript := inscription.
	//txscript.PushedData()
	builder := txscript.ScriptBuilder{}
	builder = *builder.AddFullData(pubKey.SerializeUncompressed()).AddOp(txscript.OP_CHECKSIG) // compress or un compress
	revealScript := inscript.NftRevealScript(inscription, builder)

	tapLeafSpendInfo := txscript.TapLeaf{
		LeafVersion: txscript.TaprootLeafMask,
		Script:      revealScript,
	}

	tapRootSpendInfo := txscript.AssembleTaprootScriptTree(tapLeafSpendInfo)

	ctrlBlock := tapRootSpendInfo.LeafMerkleProofs[0].ToControlBlock(pubKey)
	rootHash := tapRootSpendInfo.RootNode.TapHash()
	//outputKey := txscript.ComputeTaprootOutputKey(pubKey, rootHash[:])
	commitTxAddress, err := btcutil.NewAddressTaproot(rootHash[:], network) //???
	if err != nil {
		fmt.Println(err.Error())
	}

	_, revealFee := BuildRevealTransaction(&txscript.ControlBlock{
		InternalKey:     ctrlBlock.InternalKey,
		OutputKeyYIsOdd: ctrlBlock.OutputKeyYIsOdd,
		LeafVersion:     ctrlBlock.LeafVersion,
		InclusionProof:  ctrlBlock.InclusionProof,
	}, revealFeeRate, &wire.OutPoint{}, &wire.TxOut{
		Value:    0,
		PkScript: []byte(destination),
	}, revealScript)

	// Note: taproot address?
	unsignedCommitTx, err := BuildTransactionWithValue(*satP, inscriptions, utxos, commitTxAddress.String(), change, commitFeeRate, revealFee+TARGET_POSTAGE)
	if err != nil {
		return nil, nil, err
	}

	var output *wire.TxOut
	var vout int
	for v, txOut := range unsignedCommitTx.TxOut {
		scriptAdd := []byte(commitTxAddress.String())
		if bytes.Equal(txOut.PkScript, scriptAdd) {
			output = txOut
			vout = v
			break
		}
	}

	if output == nil {
		return nil, nil, fmt.Errorf("couldn't find tx out in unsigned commit tx")
	}

	revealTx, fee := BuildRevealTransaction(
		&ctrlBlock,
		revealFeeRate,
		&wire.OutPoint{
			Hash:  unsignedCommitTx.TxHash(),
			Index: uint32(vout),
		}, &wire.TxOut{
			PkScript: []byte(destination),
			Value:    output.Value,
		},
		revealScript,
	)
	if revealTx.TxOut[0].Value < int64(fee) {
		return nil, nil, fmt.Errorf("not enough to pay fee")
	} else {
		revealTx.TxOut[0].Value -= int64(fee)
	}

	dustLimit := mempool.GetDustThreshold(&wire.TxOut{
		PkScript: revealTx.TxOut[0].PkScript,
	})

	if revealTx.TxOut[0].Value < dustLimit {
		err := fmt.Errorf("commit transaction output would be dust")
		return nil, nil, err
	}

	prevOutputFetcher := txscript.NewCannedPrevOutputFetcher(unsignedCommitTx.TxOut[0].PkScript, unsignedCommitTx.TxOut[0].Value)
	hashCache := txscript.NewTxSigHashes(revealTx, prevOutputFetcher)
	sig, err := txscript.RawTxInTapscriptSignature(revealTx, hashCache, 0, output.Value, output.PkScript, txscript.TapLeaf{
		LeafVersion: txscript.TaprootLeafMask,
		Script:      revealScript,
	}, txscript.SigHashDefault, privKey)
	if err != nil {
		return nil, nil, err
	}

	fmt.Println(hex.EncodeToString(sig))

	return unsignedCommitTx, revealTx, nil
}

func BlockDbPath(dbType string) string {
	// The database name is based on the database type.
	dbName := "blocks_" + dbType
	if dbType == "sqlite" {
		dbName = dbName + ".db"
	}
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, "m25", dbName)
	return dbPath
}

// ffldb
func LoadBlockDB() (database.DB, error) {
	// The database name is based on the database type.
	dbType := "ffldb"
	dbPath := BlockDbPath(dbType)
	fmt.Println(dbPath)
	db, err := database.Open(dbType, dbPath, wire.TestNet3)
	if err != nil {
		// Return the error if it's not because the database doesn't
		// exist.
		if dbErr, ok := err.(database.Error); !ok || dbErr.ErrorCode !=
			database.ErrDbDoesNotExist {

			return nil, err
		}

		db, err = database.Create(dbType, dbPath, wire.TestNet3)
		if err != nil {
			return nil, err
		}
	} else {
		fmt.Println("Hello")
	}

	return db, nil
}
