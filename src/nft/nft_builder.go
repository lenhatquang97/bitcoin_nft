package nft

import (
	"bytes"
	"fmt"
	"os"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/utils"

	//"github.com/btcsuite/btcd/btcutil/schnorr/musig2"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

/*
* Have reviewed in 10/4/2023
* Remove ContentLength, BackupRecoveryKey
*
 */

const (
	MAXIMUM_BYTE           = 3 * 1024 * 1024
	PROTOCOL_TAG           = "M25"
	BODY_TAG               = "BodyM25"
	CONTENT_TYPE_TAG       = "ContentTypeM25"
	CHUNK_SIZE             = 500
	ENABLE_RBF_NO_LOCKTIME = 0xFFFFFFFD
)

type Inscription struct {
	Body        []byte
	ContentType string
}

func ParseNftFileFromTx(tx *wire.MsgTx) (*Inscription, error) {
	if len(tx.TxIn[0].Witness) == 3 {
		return ParseScriptToInscription(tx.TxIn[0].Witness[1]), nil
	}
	return nil, fmt.Errorf("not found any reveal script")
}

func NftFromTransaction(tx *wire.MsgTx) (*Inscription, error) {
	return ParseNftFileFromTx(tx)
}

/*
NftFromFile:
+ Get file from filePath
+ Get size from file
+ Get binary file
+ Get mimetype: video/mp4, audio/mp3,...
*/
func NftFromFile(filePath string) (*Inscription, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	fileSize := fileInfo.Size()
	if fileSize >= MAXIMUM_BYTE {
		return nil, fmt.Errorf("too much %d bytes for embedding NFT data", fileSize)
	}

	binFile, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	contentType, err := utils.GetFileContentType(file)
	if err != nil {
		panic(err)
	}

	return &Inscription{Body: binFile, ContentType: contentType}, nil
}

// RevealScript: use to embed file or any text into witness
func NftRevealScript(nftFile *Inscription, builder txscript.ScriptBuilder) []byte {
	builder = *builder.AddOp(txscript.OP_FALSE).AddOp(txscript.OP_IF).AddFullData([]byte(PROTOCOL_TAG))

	if len(nftFile.ContentType) != 0 && nftFile.Body != nil {
		builder = *builder.AddFullData([]byte(CONTENT_TYPE_TAG)).AddFullData([]byte(nftFile.ContentType))
		multipleChunks := utils.ChunkSlice(nftFile.Body, CHUNK_SIZE)
		for _, chunk := range multipleChunks {
			builder = *builder.AddFullData([]byte(BODY_TAG)).AddFullData(chunk)
		}
	}
	scriptVal, _ := builder.Script()
	scriptVal = append(scriptVal, txscript.OP_ENDIF)
	return scriptVal
}

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
		outpointConverted := ConvertToOutpoint(&v.PreviousOutPoint)
		sumTxIn += utxos[outpointConverted.Serialize()]
	}

	var sumTxOut int64
	for _, v := range tx.TxOut {
		sumTxOut += v.Value
	}

	return btcutil.Amount(sumTxIn - sumTxOut)
}

func CreateInscriptionTransaction(satpoint *src.SatPoint,
	inscription *Inscription,
	inscriptions map[string]src.InscriptionId,
	network *chaincfg.Params,
	utxos map[string]int64,
	change []string,
	destination string,
	commitFeeRate float64,
	revealFeeRate float64,
	noLimit bool,
) (*wire.MsgTx, *wire.MsgTx, error) {
	var satP *src.SatPoint
	if satpoint != nil {
		satP = satpoint
	} else {
		inscribeUtxos := make(map[string]string) // find about set in golang
		for inscrp := range inscriptions {
			s, err := src.DeserializeSatPoint(inscrp)
			if err != nil {
				return nil, nil, err
			}

			inscribeUtxos[s.OutPoint.Serialize()] = ""
		}

		for outpoint := range utxos {
			_, ok := inscribeUtxos[outpoint]
			if !ok {
				outpointConverted, err := DeserializeOutpoint(outpoint)
				if err != nil {
					return nil, nil, err
				}

				satP = &src.SatPoint{
					OutPoint: *outpointConverted,
					OffSet:   0,
				}
				break
			}
		}
	}

	satpointSerialize := satpoint.Serialize()
	for inscribedSatpoint, inscriptionId := range inscriptions {
		if inscribedSatpoint == satpointSerialize {
			return nil, nil, fmt.Errorf("sat at %v sat poiont already inscribed", satpoint)
		}

		satpointDeserialize, err := src.DeserializeSatPoint(inscribedSatpoint)
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
	revealScript := NftRevealScript(inscription, builder)

	tapLeafSpendInfo := txscript.TapLeaf{
		LeafVersion: txscript.TaprootLeafMask,
		Script:      revealScript,
	}

	tapRootSpendInfo := txscript.AssembleTaprootScriptTree(tapLeafSpendInfo)

	ctrlBlock := tapRootSpendInfo.LeafMerkleProofs[0].ToControlBlock(pubKey)
	rootHash := tapRootSpendInfo.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(pubKey, rootHash[:])
	commitTxAddress, _ := btcutil.NewAddressTaproot(outputKey.SerializeUncompressed(), network) //???

	_, revealFee := BuildRevealTransaction(&txscript.ControlBlock{
		InternalKey:     ctrlBlock.InternalKey,
		OutputKeyYIsOdd: ctrlBlock.OutputKeyYIsOdd,
		LeafVersion:     ctrlBlock.LeafVersion,
		InclusionProof:  ctrlBlock.InclusionProof,
	}, revealFeeRate, nil, &wire.TxOut{
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
		if bytes.Equal(output.PkScript, commitTxAddress.ScriptAddress()) {
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

	hashCache := txscript.NewTxSigHashes(revealTx, blockchain.NewUtxoViewpoint())
	sig, err := txscript.RawTxInTapscriptSignature(revealTx, hashCache, 0, output.Value, output.PkScript, txscript.TapLeaf{
		LeafVersion: txscript.TaprootLeafMask,
		Script:      revealScript,
	}, txscript.SigHashDefault, privKey)

	if err != nil {
		return nil, nil, err
	}

	fmt.Println(sig)
	return unsignedCommitTx, revealTx, nil
}

/*
* ParseScriptToInscription: Convert Reveal Script into Inscription
 */
func ParseScriptToInscription(script []byte) *Inscription {
	result := Inscription{}
	startSomeOpcode := 1 + 1 + utils.GetPaddingInAddData([]byte(PROTOCOL_TAG)) + len([]byte(PROTOCOL_TAG))
	startContentType := startSomeOpcode + utils.GetPaddingInAddData([]byte(CONTENT_TYPE_TAG)) + len([]byte(CONTENT_TYPE_TAG)) + 1
	endContentType := utils.GetPaddingInAddData([]byte(BODY_TAG)) + utils.FindAPartOfByteArray([]byte(BODY_TAG), script) - 1
	result.ContentType = string(script[startContentType:endContentType])

	multipleIndexes := utils.FindMultiplePartsOfByteArray([]byte(BODY_TAG), script)
	for i := 0; i < len(multipleIndexes)-1; i++ {
		startChunkWithPadding := multipleIndexes[i] + len([]byte(BODY_TAG))
		endChunk := multipleIndexes[i+1] - utils.GetPaddingInAddData([]byte(BODY_TAG))
		padding := utils.GetPaddingInAddData(script[startChunkWithPadding:endChunk])
		actualStartChunk := startChunkWithPadding + padding
		actualEndChunk := endChunk
		result.Body = append(result.Body, script[actualStartChunk:actualEndChunk]...)
	}

	startChunkWithPadding := multipleIndexes[len(multipleIndexes)-1] + len([]byte(BODY_TAG))
	endChunk := len(script) - 1
	padding := utils.GetPaddingInAddData(script[startChunkWithPadding:endChunk])

	actualStartBody := startChunkWithPadding + padding
	actualEndBody := endChunk
	result.Body = append(result.Body, script[actualStartBody:actualEndBody]...)

	return &result
}
