package nft

import (
	"fmt"
	"log"
	"os"

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

type NftFile struct {
	Body        []byte
	ContentType string
}

// TODO: Retrieve NftFile from Transaction
func ParseNftFileFromTx(tx *wire.MsgTx) {

}

/*
NftFromFile:
+ Get file from filePath
+ Get size from file
+ Get binary file
+ Get mimetype: video/mp4, audio/mp3,...
*/
func NftFromFile(filePath string) *NftFile {
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

	return &NftFile{Body: binFile, ContentType: contentType}
}

// Reveal Script
func NftRevealScriptBuilder(nftFile *NftFile) *txscript.ScriptBuilder {
	builder := txscript.ScriptBuilder{}
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

func BuildRevealTransaction(ctrlBlock *txscript.ControlBlock, feeRate float64, input *wire.OutPoint, output *wire.TxOut, script []byte) (*wire.MsgTx, int) {
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
	actualFee := 1000
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
