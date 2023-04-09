package nft

import (
	"fmt"
	"io/ioutil"
	"os"

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
func FindAPartOfByteArray(part []byte, array []byte) int {
	m := len(part)
	n := len(array)

	for i := 0; i <= n-m; i++ {
		var j = 0

		for j = 0; j < m; j++ {
			if array[i+j] != part[j] {
				break
			}
		}
		if j == m {
			return i
		}
	}
	return -1
}

func FindMultiplePartsOfByteArray(part []byte, array []byte) []int {
	m := len(part)
	n := len(array)

	result := make([]int, 0)

	for i := 0; i <= n-m; i++ {
		var j = 0
		for j = 0; j < m; j++ {
			if array[i+j] != part[j] {
				break
			}
		}
		if j == m {
			result = append(result, i)
		}
	}
	return result
}

func GetPadding(data []byte) int {
	dataLen := len(data)
	//Case it's a opcode
	condition1 := dataLen == 0 || dataLen == 1 && data[0] == 0
	condition2 := dataLen == 1 && data[0] <= 16
	condition3 := dataLen == 1 && data[0] == 0x81
	if condition1 || condition2 || condition3 {
		return 1
	}

	if dataLen < txscript.OP_PUSHDATA1 {
		return 1
	} else if dataLen <= 0xff {
		return 2
	} else if dataLen <= 0xffff {
		return 3
	} else {
		return 5
	}
}

func ParseScriptToInscription(script []byte) *Inscription {
	result := Inscription{}
	startSomeOpcode := 1 + 1 + GetPadding([]byte(PROTOCOL_TAG)) + len([]byte(PROTOCOL_TAG))
	startContentType := startSomeOpcode + GetPadding([]byte(CONTENT_TYPE_TAG)) + len([]byte(CONTENT_TYPE_TAG)) + 1
	endContentType := GetPadding([]byte(BODY_TAG)) + FindAPartOfByteArray([]byte(BODY_TAG), script) - 1
	result.ContentType = string(script[startContentType:endContentType])

	multipleIndexes := FindMultiplePartsOfByteArray([]byte(BODY_TAG), script)
	for i := 0; i < len(multipleIndexes)-1; i++ {
		startChunkWithPadding := multipleIndexes[i] + len([]byte(BODY_TAG))
		endChunk := multipleIndexes[i+1] - GetPadding([]byte(BODY_TAG))
		padding := GetPadding(script[startChunkWithPadding:endChunk])
		actualStartChunk := startChunkWithPadding + padding
		actualEndChunk := endChunk
		result.Body = append(result.Body, script[actualStartChunk:actualEndChunk]...)
	}

	startChunkWithPadding := multipleIndexes[len(multipleIndexes)-1] + len([]byte(BODY_TAG))
	endChunk := len(script) - 1
	padding := GetPadding(script[startChunkWithPadding:endChunk])

	actualStartBody := startChunkWithPadding + padding
	actualEndBody := endChunk
	result.Body = append(result.Body, script[actualStartBody:actualEndBody]...)

	return &result
}

func TestWithSimplePDF() {
	inscription, _ := NftFromFile("./Consensus4.pdf")
	script := NftRevealScript(inscription, *txscript.NewScriptBuilder())
	result := ParseScriptToInscription(script)
	err := os.WriteFile("./final.pdf", result.Body, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result.ContentType)
}

func TestWithSimpleText() {
	inscription := Inscription{
		ContentType: "text/plain",
		Body:        []byte("Hello World"),
	}
	script := NftRevealScript(&inscription, *txscript.NewScriptBuilder())
	result := ParseScriptToInscription(script)
	fmt.Println(len(string(result.Body)))
}
