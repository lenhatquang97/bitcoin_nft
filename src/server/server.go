package main

import (
	"github.com/gin-gonic/gin"
	"github.com/m25lab/bitcoin_nft/src/nft"
	"log"
	"os"
)

type Search struct {
	Query string
}

var server *gin.Engine

type Server struct {
	Address             string
	AcmeDomain          []string // no use
	HttpPort            int64
	HttpsPort           int64
	AcmeCache           string
	AcmeContact         []string
	Http                bool
	Https               bool
	RedirectHttpToHttps bool
}

func init() {

	server = gin.Default()
}

func run(opt *nft.Options, index nft.Index) {
	newIndex := index
	nft.Update(&newIndex)

	//config, err := nft.LoadConfig(opt)
	//if err != nil {
	//	return err
	//}

}

func main() {
	basePath := server.Group("/v1")
	RegisterRoutePath(basePath)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Fatal(server.Run(":" + port))
}
