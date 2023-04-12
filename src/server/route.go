package main

import "github.com/gin-gonic/gin"

func RegisterRoutePath(rg *gin.RouterGroup) {
	nft := rg.Group("/nft")
	nft.GET("/test", Test)
}
