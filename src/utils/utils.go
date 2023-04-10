package utils

import (
	"math/big"
	"net/http"
	"os"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/txscript"
)

type Account struct {
	Address btcutil.Address
	Amount  btcutil.Amount
}

func IntToBytes(i int) []byte {
	if i > 0 {
		return append(big.NewInt(int64(i)).Bytes(), byte(1))
	}
	return append(big.NewInt(int64(i)).Bytes(), byte(0))
}

func BytesToInt(b []byte) int64 {
	if b[len(b)-1] == 0 {
		return -big.NewInt(0).SetBytes(b[:len(b)-1]).Int64()
	}
	return big.NewInt(0).SetBytes(b[:len(b)-1]).Int64()
}

func GetFileContentType(out *os.File) (string, error) {
	buffer := make([]byte, 512)
	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}
func ChunkSlice(slice []byte, chunkSize int) [][]byte {
	var chunks [][]byte
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}

		chunks = append(chunks, slice[i:end])
	}

	return chunks
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

func GetPaddingInAddData(data []byte) int {
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
