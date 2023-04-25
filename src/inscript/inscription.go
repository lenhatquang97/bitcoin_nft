package inscript

import (
	"fmt"
	"os"

	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/m25lab/bitcoin_nft/src/model"
	"github.com/m25lab/bitcoin_nft/src/utils"
)

const (
	MAXIMUM_BYTE           = 3 * 1024 * 1024
	PROTOCOL_TAG           = "M25"
	BODY_TAG               = "BodyM25"
	CONTENT_TYPE_TAG       = "ContentTypeM25"
	CHUNK_SIZE             = 500
	ENABLE_RBF_NO_LOCKTIME = 0xFFFFFFFD
)

/*
NftFromFile:
+ Get file from filePath
+ Get size from file
+ Get binary file
+ Get mimetype: video/mp4, audio/mp3,...
*/
func NftFromFile(filePath string) (*model.Inscription, error) {
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

	return &model.Inscription{Body: binFile, ContentType: contentType}, nil
}

// RevealScript: use to embed file or any text into witness
func NftRevealScript(nftFile *model.Inscription, builder txscript.ScriptBuilder) []byte {
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

func ParseNftFileFromTx(tx *wire.MsgTx) (*model.Inscription, error) {
	if len(tx.TxIn[0].Witness) == 3 {
		return ParseScriptToInscription(tx.TxIn[0].Witness[1]), nil
	}
	return nil, fmt.Errorf("not found any reveal script")
}

func NftFromTransaction(tx *wire.MsgTx) (*model.Inscription, error) {
	return ParseNftFileFromTx(tx)
}

func ParseScriptToInscription(script []byte) *model.Inscription {
	result := model.Inscription{}
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
