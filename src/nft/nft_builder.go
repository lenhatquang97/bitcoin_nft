package nft

import (
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcec/v2/schnorr/musig2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/mempool"
	"github.com/m25lab/bitcoin_nft/src"
	"log"
	"os"
	//"github.com/btcsuite/btcd/btcutil/schnorr/musig2"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const (
	MAXIMUM_BYTE     = 3 * 1024 * 1024
	PROTOCOL_TAG     = "M25"
	BODY_TAG         = "BodyM25"
	CONTENT_TYPE_TAG = "ContentTypeM25"
	CHUNK_SIZE       = MAXIMUM_BYTE / 6
)

type Inscription struct {
	Body        []byte
	ContentType string
}

// TODO: Retrieve Inscription from Transaction
func ParseNftFileFromTx(tx *wire.MsgTx) {

}

/*
NftFromFile:
+ Get file from filePath
+ Get size from file
+ Get binary file
+ Get mimetype: video/mp4, audio/mp3,...
*/
func NftFromFile(filePath string) *Inscription {
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}

	fileSize := fileInfo.Size()
	if fileSize >= MAXIMUM_BYTE {
		panic(fmt.Sprintf("Too much %d bytes for embedding NFT data", fileSize))
	}

	binFile, err := os.ReadFile(filePath)
	if err != nil {
		panic(err)
	}

	contentType, err := GetFileContentType(file)
	if err != nil {
		panic(err)
	}

	return &Inscription{Body: binFile, ContentType: contentType}
}

// Reveal Script
func NftRevealScriptBuilder(nftFile *Inscription, builder txscript.ScriptBuilder) *txscript.ScriptBuilder {
	builder = *builder.AddOp(txscript.OP_FALSE).AddOp(txscript.OP_IF).AddData([]byte(PROTOCOL_TAG))

	if len(nftFile.ContentType) != 0 && nftFile.Body != nil {
		builder = *builder.AddData([]byte(CONTENT_TYPE_TAG)).AddData([]byte(nftFile.ContentType))
		multipleChunks := ChunkSlice(nftFile.Body, CHUNK_SIZE)
		for _, chunk := range multipleChunks {
			builder = *builder.AddData([]byte(BODY_TAG)).AddData(chunk)
		}
	}
	builder = *builder.AddOp(txscript.OP_ENDIF)
	return &builder
}

func NftRevealScript(nftFile *Inscription, builder txscript.ScriptBuilder) ([]byte, error) {
	return NftRevealScriptBuilder(nftFile, builder).Script()
}

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

	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, []byte{0, schnorr.SignatureSize})

	ctrlBlockByte, err := ctrlBlock.ToBytes()
	if err != nil {
		log.Println(err)
	}
	copyTx.TxIn[0].Witness = append(copyTx.TxIn[0].Witness, ctrlBlockByte)
	txSize := mempool.GetTxVirtualSize(btcutil.NewTx(copyTx))

	actualFee := btcutil.Amount(feeRate * float64(txSize))

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
) (*wire.MsgTx, *wire.MsgTx, *musig2.KeyTweakDesc, error) {
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
	revealScript, err := NftRevealScript(inscription, builder)

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

	unsignedCommitTx :=
}
