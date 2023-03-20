package main

import (
	"fmt"

	"github.com/m25lab/bitcoin_nft/src/initial"
)

func main() {
	// _, addr := initial.GenerateBTCAddress(initial.P2TR)
	// fmt.Println(addr)
	loadedConfig := initial.ReadCAFile()
	fmt.Println(loadedConfig)
}
