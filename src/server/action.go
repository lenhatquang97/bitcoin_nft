package main

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gin-gonic/gin"
	"github.com/m25lab/bitcoin_nft/src"
	"github.com/m25lab/bitcoin_nft/src/nft"
)

func (sv *Server) Test(ctx *gin.Context) {

}

// maybe not need impl
func (sv *Server) Home(ctx *gin.Context) {
	type HomeResponse struct {
		Last           int64
		Blocks         string
		InscriptionIds []src.InscriptionId
	}

	_ = nft.GetBlock(sv.Index)

}

// maybe not need impl
func (sv *Server) BlockCount(ctx *gin.Context) {

}

// get block
func (sv *Server) Block(c *gin.Context) {
	type BlockHashInfo struct {
		Height int64          `json:"height"`
		Hash   chainhash.Hash `json:"hash"`
	}

	type BlockInfo struct {
		Height int64          `json:"height"`
		Block  *wire.MsgBlock `json:"block"`
	}

	q, ok := c.GetQuery("q")
	if !ok {
		c.JSON(404, "Param is empty")
	}

	var input *BlockHashInfo
	err := json.Unmarshal([]byte(q), &input)
	if err != nil {
		c.JSON(500, err)
	}

	if input.Height > 0 {
		res, err := nft.GetBlockByHeight(sv.Index, input.Height)
		if err != nil {
			c.JSON(500, err)
		}

		c.JSON(200, &BlockInfo{
			Height: input.Height,
			Block:  res,
		})
	} else if input.Hash.String() != "" {
		block, err := nft.GetBlockByHash(sv.Index, &input.Hash)
		if err != nil {
			c.JSON(500, err)
		}

		//res, err := nft.BlockHeader(sv.Index, &input.Hash)
		//if err != nil {
		//	c.JSON(500, err)
		//}

		c.JSON(200, &BlockInfo{
			Block: block,
		})
	} else {
		c.JSON(500, "Input is invalid")
	}
}

// no need to impl
func (sv *Server) Bounties(ctx *gin.Context) {

}

// Get block height in db (not impl)
func (sv *Server) Clock(ctx *gin.Context) {

}

func (sv *Server) Content(c *gin.Context) {
	q, ok := c.GetQuery("q")
	if !ok {
		c.JSON(404, "Param is nil")
	}

	var input *src.InscriptionId
	err := json.Unmarshal([]byte(q), input)
	if err != nil {
		c.JSON(500, err)
	}

	res, err := nft.GetInscriptionById(sv.Index, input)
	if err != nil {
		c.JSON(500, err)
	}

	c.JSON(200, res)
}

// no need to impl
func (sv *Server) Faq(ctx *gin.Context) {

}

func (sv *Server) TransactionInput(c *gin.Context) {
	type Input struct {
		Height      int64 `json:"height"`
		Transaction int64 `json:"transaction"`
		TxIn        int64 `json:"txIn"`
	}

	q, ok := c.GetQuery("q")
	if !ok {
		c.JSON(500, "Param is nil")
	}

	var input *Input
	err := json.Unmarshal([]byte(q), &input)
	if err != nil {
		c.JSON(500, err)
	}

	block, err := nft.GetBlockByHeight(sv.Index, input.Height)
	if err != nil {
		c.JSON(500, err)
	}

	if input.Transaction > int64(len(block.Transactions)) {
		c.JSON(500, fmt.Sprintf("Transaction index is out of range of tx data %v - %v", input.Transaction, len(block.Transactions)))
	}
	transaction := block.Transactions[input.Transaction]

	if int64(len(transaction.TxIn)) < input.TxIn {
		c.JSON(500, fmt.Sprintf("Transaction index is out of range of tx data %v - %v", input.TxIn, len(transaction.TxIn)))
	}

	txIn := transaction.TxIn[input.TxIn]

	c.JSON(200, txIn)
}

// no need to get prev and next inscription
func (sv *Server) Inscription(c *gin.Context) {
	q, ok := c.GetQuery("q")
	if !ok {
		c.JSON(404, "Param is empty")
	}

	var input *src.InscriptionId
	err := json.Unmarshal([]byte(q), &input)
	if err != nil {
		c.JSON(500, err)
	}

	inscription, err := nft.GetInscriptionById(sv.Index, input)
	if err != nil {
		c.JSON(500, err)
	}

	satPoint, err := nft.GetInscriptionSatPointById(sv.Index, input)
	if err != nil {
		c.JSON(500, err)
	}

	txId := satPoint.OutPoint.TxidStr
	output, err := nft.GetTransaction(sv.Index, txId)
	if err != nil {
		c.JSON(500, err)
	}

	type response struct {
		Inscription *nft.Inscription `json:"inscription"`
		Output      *wire.MsgTx      `json:"output"`
	}

	c.JSON(200, response{
		Inscription: inscription,
		Output:      output,
	})
}

// not yet need to impl
func (sv *Server) InscriptionList(ctx *gin.Context) {

}

// not yet need to impl
func (sv *Server) InscriptionFrom(ctx *gin.Context) {

}

// not yet need to impl (just redirect)
func (sv *Server) OrdSat(ctx *gin.Context) {

}

// no need to config sat index
func (sv *Server) Output(c *gin.Context) {
	q, ok := c.GetQuery("q")
	if !ok {
		c.JSON(500, "Param input is nil")
	}

	var input nft.Outpoint
	err := json.Unmarshal([]byte(q), &input)
	if err != nil {
		c.JSON(500, err)
	}

	txId := input.TxidStr
	tx, err := nft.GetTransaction(sv.Index, txId)
	if err != nil {
		c.JSON(500, err)
	}

	txOut := tx.TxOut[input.OutputIndex]
	inscriptionIds, err := nft.GetInscriptionOnOutput(sv.Index, input)
	if err != nil {
		c.JSON(500, err)
	}

	type response struct {
		TxOut          *wire.TxOut         `json:"txOut"`
		InscriptionIds []src.InscriptionId `json:"inscriptionIds"`
	}

	c.JSON(200, response{
		TxOut:          txOut,
		InscriptionIds: inscriptionIds,
	})
}

// High priority: no need, just dispplay inscription to UI
func (sv *Server) PreviewInscription(ctx *gin.Context) {

}

// Low prioriry
func (sv *Server) Range(ctx *gin.Context) {

}

// Get rate satpoint
func (sv *Server) Sat(ctx *gin.Context) {

}

// no need to impl
func (sv *Server) SearchByQuery(ctx *gin.Context) {

}

// no need to impl
func (sv *Server) SearchByPath(ctx *gin.Context) {

}

// no need to impl
func (sv *Server) StaticAsset(ctx *gin.Context) {

}

// no need to impl
func (sv *Server) Status(ctx *gin.Context) {

}

// High priority
func (sv *Server) Transaction(c *gin.Context) {
	txId, ok := c.GetQuery("q")
	if !ok {
		c.JSON(500, "Param is nil")
	}

	inscription, err := nft.GetInscriptionById(sv.Index, &src.InscriptionId{
		TxID: txId,
	})

	if err != nil {
		c.JSON(500, err)
	}

	blockHash, err := nft.GetTransactionBlockHash(sv.Index, txId)
	if err != nil {
		c.JSON(500, err)
	}

	type response struct {
		Hash        *chainhash.Hash  `json:"hash"`
		Inscription *nft.Inscription `json:"inscription"`
	}

	c.JSON(200, response{
		Inscription: inscription,
		Hash:        blockHash,
	})
}
