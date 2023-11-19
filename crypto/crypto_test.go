package crypto

import (
	"dpchain/common"
	"encoding/hex"
	"testing"
)

func TestSignatureValid(t *testing.T) {
	// 你需要创建一个测试用的地址、签名和哈希
	// 用于测试 SignatureValid 函数

	// 示例地址
	address := common.HexToAddress("5789f96b2a695001646acd26068725ee0933fa06")

	// 示例哈希
	hash := common.HexToHash("df68ae24ef8745f5641333fff3981df4e742ff106da94fc7e0cb044a944f4aa0")

	hexSignature := "321961eb323cedbb5b0388493edc0a94b5b891eb4d2e65e50e536885bcf52c545c86362794b42f9720b193cbb1afbfcc771a5f9661277237548820ff456cfe3300"

	// 将十六进制字符串转换为字节切片

	// 示例签名
	signature := []byte("321961eb323cedbb5b0388493edc0a94b5b891eb4d2e65e50e536885bcf52c545c86362794b42f9720b193cbb1afbfcc771a5f9661277237548820ff456cfe3300")

	signature, err := hex.DecodeString(hexSignature)
	valid, err := SignatureValid(address, signature, hash)

	if err != nil {
		t.Errorf("Error while verifying signature: %v", err)
	}

	if valid {
		t.Logf("Signature is valid for the provided address.")
	} else {
		t.Errorf("Signature is not valid for the provided address.")
	}
}

func main() {
	// 运行测试
	testing.Main(func(pat, str string) (bool, error) { return true, nil }, []testing.InternalTest{
		{F: TestSignatureValid},
	}, nil, nil)
}
