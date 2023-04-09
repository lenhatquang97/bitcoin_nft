package nft

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/btcsuite/btcd/blockchain"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/m25lab/bitcoin_nft/src"

	//"github.com/btcsuite/btcd/btcutil/schnorr/musig2"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

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

// TODO: Retrieve Inscription from Transaction
// impl soon (1) (Quang)
func ParseNftFileFromTx(tx *wire.MsgTx) (*Inscription, error) {
	return nil, nil
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
		return nil, errors.New(fmt.Sprintf("Too much %d bytes for embedding NFT data", fileSize))
	}

	binFile, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	contentType, err := GetFileContentType(file)
	if err != nil {
		panic(err)
	}

	return &Inscription{Body: binFile, ContentType: contentType}, nil
}

func NftFromTransaction(tx *wire.MsgTx) (*Inscription, error) {
	return ParseNftFileFromTx(tx)
}

// Reveal Script --> reverse?
func NftRevealScript(nftFile *Inscription, builder txscript.ScriptBuilder) []byte {
	scriptVal, err := builder.Script()
	if err != nil {
		fmt.Println(err)
	}

	scriptVal = append(scriptVal, txscript.OP_FALSE)
	scriptVal = append(scriptVal, txscript.OP_IF)
	scriptVal = append(scriptVal, []byte(PROTOCOL_TAG)...)

	if len(nftFile.ContentType) != 0 && nftFile.Body != nil {
		scriptVal = append(scriptVal, []byte(CONTENT_TYPE_TAG)...)
		scriptVal = append(scriptVal, []byte(nftFile.ContentType)...)

		multipleChunks := ChunkSlice(nftFile.Body, CHUNK_SIZE)
		for _, chunk := range multipleChunks {
			scriptVal = append(scriptVal, []byte(BODY_TAG)...)
			scriptVal = append(scriptVal, chunk...)
		}
	}

	builder = *builder.AddOp(txscript.OP_FALSE).AddOp(txscript.OP_IF).AddFullData([]byte(PROTOCOL_TAG))

	if len(nftFile.ContentType) != 0 && nftFile.Body != nil {
		builder = *builder.AddFullData([]byte(CONTENT_TYPE_TAG)).AddFullData([]byte(nftFile.ContentType))
		multipleChunks := ChunkSlice(nftFile.Body, CHUNK_SIZE)
		for _, chunk := range multipleChunks {
			builder = *builder.AddFullData([]byte(BODY_TAG)).AddFullData(chunk)
		}
	}
	scriptVal, _ = builder.Script()
	scriptVal = append(scriptVal, txscript.OP_ENDIF)
	return scriptVal
}

// Check lại thiếu param script
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

	//TODO: Need to calculate fee
	copyTx := revealTx.Copy()

	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, make([]byte, schnorr.SignatureSize))

	ctrlBlockByte, err := ctrlBlock.ToBytes()
	if err != nil {
		log.Println(err)
	}
	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, ctrlBlockByte)
	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(copyTx))

	actualFee := Fee(feeRate, float64(txSize))

	return &revealTx, actualFee
}

func ToWitness(builder *txscript.ScriptBuilder) wire.TxWitness {
	script, err := builder.Script()
	if err != nil {
		log.Fatal(err)
	}
	witness := wire.TxWitness(make([][]byte, 2))
	witness[0] = script
	witness[1] = make([]byte, 0)
	return witness
}

func CalculateFee(tx *wire.MsgTx, utxos map[wire.OutPoint]btcutil.Amount) btcutil.Amount {
	var sumTxIn btcutil.Amount
	for _, v := range tx.TxIn {
		sumTxIn += utxos[v.PreviousOutPoint]
	}

	var sumTxOut int64
	for _, v := range tx.TxOut {
		sumTxOut += v.Value
	}

	return sumTxIn - btcutil.Amount(sumTxOut)
}

func CreateInscriptionTransaction(satpoint *src.SatPoint,
	inscription *Inscription,
	inscriptions map[src.SatPoint]src.InscriptionId,
	network *chaincfg.Params,
	utxos map[wire.OutPoint]btcutil.Amount,
	change []btcutil.Address,
	destination btcutil.Address,
	commitFeeRate float64,
	revealFeeRate float64,
	noLimit bool,
) (*wire.MsgTx, *wire.MsgTx, *btcec.PrivateKey, error) {
	var satP *src.SatPoint
	if satpoint != nil {
		satP = satpoint
	} else {
		inscribeUtxos := make(map[wire.OutPoint]string) // find about set in golang
		for inscrp := range inscriptions {
			inscribeUtxos[inscrp.OutPoint] = ""
		}
		for outpoint := range utxos {
			_, ok := inscribeUtxos[outpoint]
			if !ok {
				satP = &src.SatPoint{
					OutPoint: outpoint,
					OffSet:   0,
				}
				break
			}
		}
	}

	for inscribedSatpoint, inscriptionId := range inscriptions {
		if satpoint == nil {
			continue
		}

		if inscribedSatpoint == *satpoint {
			return nil, nil, nil, errors.New(fmt.Sprintf("Sat at %v sat poiont already inscribed", satpoint))
		}

		if inscribedSatpoint.OutPoint == satpoint.OutPoint {
			return nil, nil, nil, errors.New(fmt.Sprintf("utxo already inscribed %v on sat %v", inscriptionId, inscribedSatpoint))
		}
	}

	privKey, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, nil, nil, err
	}

	pubKey := privKey.PubKey()
	//revealScript := inscription.
	//txscript.PushedData()
	builder := txscript.ScriptBuilder{}
	builder = *builder.AddData(pubKey.SerializeUncompressed()).AddOp(txscript.OP_CHECKSIG) // compress or un compress
	revealScript := NftRevealScript(inscription, builder)

	tapLeafSpendInfo := txscript.TapLeaf{
		LeafVersion: txscript.TaprootLeafMask,
		Script:      revealScript,
	}

	tapRootSpendInfo := txscript.AssembleTaprootScriptTree(tapLeafSpendInfo)
	//ctrlBlock :=

	ctrlBlock := tapRootSpendInfo.LeafMerkleProofs[0].ToControlBlock(pubKey)
	rootHash := tapRootSpendInfo.RootNode.TapHash()
	outputKey := txscript.ComputeTaprootOutputKey(pubKey, rootHash[:])
	commitTxAddress, err := btcutil.NewAddressTaproot(outputKey.SerializeUncompressed(), network)

	_, revealFee := BuildRevealTransaction(&txscript.ControlBlock{
		InternalKey:     ctrlBlock.InternalKey,
		OutputKeyYIsOdd: ctrlBlock.OutputKeyYIsOdd,
		LeafVersion:     ctrlBlock.LeafVersion,
		InclusionProof:  ctrlBlock.InclusionProof,
	}, revealFeeRate, nil, &wire.TxOut{
		Value:    0,
		PkScript: destination.ScriptAddress(),
	}, revealScript)

	unsignedCommitTx, err := BuildTransactionWithValue(*satP, inscriptions, utxos, commitTxAddress, change, commitFeeRate, revealFee+TARGET_POSTAGE)
	if err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, errors.New("Couldn't find tx out in unsigned commit tx")
	}

	revealTx, fee := BuildRevealTransaction(
		&ctrlBlock,
		revealFeeRate,
		&wire.OutPoint{
			Hash:  unsignedCommitTx.TxHash(),
			Index: uint32(vout),
		}, &wire.TxOut{
			PkScript: destination.ScriptAddress(),
			Value:    output.Value,
		},
		revealScript,
	)

	revealTx.TxOut[0].Value -= int64(fee)

	dustLimit := mempool.GetDustThreshold(&wire.TxOut{
		PkScript: revealTx.TxOut[0].PkScript,
	})

	if revealTx.TxOut[0].Value < dustLimit {
		err := errors.New(fmt.Sprintf("Commit transaction output would be dust!"))
		return nil, nil, nil, err
	}

	hashCache := txscript.NewTxSigHashes(revealTx, blockchain.NewUtxoViewpoint())
	// amt not yet set value

	// private key or compare with privTweak?
	sig, err := txscript.RawTxInTapscriptSignature(revealTx, hashCache, 0, output.Value, output.PkScript, txscript.TapLeaf{
		LeafVersion: txscript.TaprootLeafMask,
		Script:      revealScript,
	}, txscript.SigHashDefault, privKey)

	if err != nil {
		return nil, nil, nil, err
	}

	witness := revealTx.TxIn[0].Witness
	ctrlBlockByte, err := ctrlBlock.ToBytes()
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, err
	}
	witness = append(witness, sig)
	witness = append(witness, revealScript)
	witness = append(witness, ctrlBlockByte)

	tweakPrivKey := txscript.TweakTaprootPrivKey(*privKey, rootHash[:])

	// check only public key
	// check weight

	return unsignedCommitTx, revealTx, tweakPrivKey, nil
}

func BackupRecoverKey() {

}

func ContentLength(inscription *Inscription) int {
	return len(inscription.Body)
}

// 1 2 3 4 5
// 2 3 4
func FindFirstPartOfByte(script []byte, searchKey []byte) int {
	m := len(searchKey)
	n := len(script)

	for i := 0; i <= n-m; i++ {
		var j = 0
		for j = 0; j < m; j++ {
			if script[i+j] != searchKey[j] {
				break
			}
		}
		if j == m {
			return i
		}
	}
	return -1
}

func FindMultipleStartOfByte(script []byte, searchKey []byte) []int {
	m := len(searchKey)
	n := len(script)

	var result []int

	for i := 0; i <= n-m; i++ {
		var j = 0
		for j = 0; j < m; j++ {
			if script[i+j] != searchKey[j] {
				break
			}
		}
		if j == m {
			result = append(result, i)
		}
	}
	return result
}

func AddPadding(data []byte) int {
	padding := 0
	dataLen := len(data)
	condition1 := dataLen == 0 || dataLen == 1 && data[0] == 0
	condition2 := dataLen == 1 && data[0] <= 16
	condition3 := dataLen == 1 && data[0] == 0x81

	if condition1 || condition2 || condition3 {
		return dataLen
	}

	if dataLen < txscript.OP_PUSHDATA1 {
		padding += 1
	} else if dataLen <= 0xff {
		padding += 2
	} else if dataLen <= 0xffff {
		padding += 3
	} else {
		padding += 5
	}
	return padding
}

func ConvertNftRevealScript(script []byte) *Inscription {
	result := Inscription{}
	startIndex := 1 + 1 + len([]byte(PROTOCOL_TAG))

	//[contentTypeStartIndex, contentTypeLastIndex)
	whetherHasContentType := FindFirstPartOfByte(script, []byte(CONTENT_TYPE_TAG))
	if whetherHasContentType >= 0 {
		contentTypeStartIndex := startIndex + len([]byte(CONTENT_TYPE_TAG))
		contentTypeLastIndex := FindFirstPartOfByte(script, []byte(BODY_TAG))
		contentType := string(script[contentTypeStartIndex:contentTypeLastIndex])
		result.ContentType = contentType
		//[start, end)
		multipleStart := FindMultipleStartOfByte(script, []byte(BODY_TAG))
		fullBinary := []byte{}
		for i := 0; i < len(multipleStart)-1; i++ {
			startBodyIndex := multipleStart[i] + len([]byte(BODY_TAG))
			endBodyIndex := multipleStart[i+1]
			fullBinary = append(fullBinary, script[startBodyIndex:endBodyIndex]...)
		}
		startBodyIndex := multipleStart[len(multipleStart)-1] + len([]byte(BODY_TAG))
		endBodyIndex := len(script) - 1

		fullBinary = append(fullBinary, script[startBodyIndex:endBodyIndex]...)
		result.Body = fullBinary
	}
	return &result
}
