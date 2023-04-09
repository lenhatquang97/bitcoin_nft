package nft

import (
	"fmt"
	"io/ioutil"

	"github.com/btcsuite/btcd/txscript"
)

func Test1WithSimpleInscription() {
	inscription := Inscription{
		ContentType: "text/plain;charset=utf-8",
		Body:        []byte("Hello World.It's me Mario,Time!"),
	}
	script := NftRevealScript(&inscription, *txscript.NewScriptBuilder())
	finalInscription := ConvertNftRevealScript(script)
	fmt.Println(finalInscription.ContentType)
	fmt.Println(string(finalInscription.Body))
}

func Test2WithSimpleFile() {
	inscription, err := NftFromFile("./500kb.png")
	if err != nil {
		fmt.Println(err)
	}

	script := NftRevealScript(inscription, *txscript.NewScriptBuilder())
	finalInscription := ConvertNftRevealScript(script)
	fmt.Println(finalInscription.ContentType)
	//Write byte array to file named hello.png with Golang
	err = ioutil.WriteFile("file.png", finalInscription.Body, 0644)
	if err != nil {
		fmt.Println(err)
	}

}
