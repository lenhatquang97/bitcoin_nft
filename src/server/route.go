package main

import "github.com/gin-gonic/gin"

func RegisterRoutePath(rg *gin.RouterGroup, sv *Server) {
	nft := rg.Group("/nft")
	nft.GET("/test", sv.Test)
	// home 			: 		not impl
	// block count 		:		not impl
	// bounties			: 		not impl
	// clock			: 		not impl
	nft.GET("/content", sv.Content)
	nft.GET("/input", sv.TransactionInput)
	nft.GET("/inscription", sv.Inscription)
	nft.GET("/output", sv.Output)
	nft.GET("/tx", sv.Transaction)
}
