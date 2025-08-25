//go:build !js
// +build !js

package save

import (
	"fmt"
	"os"
)

func GetWallets() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var walls = []string{}
	for _, v := range entries {
		if v.IsDir() || len(v.Name()) < 6 {
			continue
		}
		if v.Name()[len(v.Name())-5:] != ".keys" {
			continue
		}
		n := v.Name()[:len(v.Name())-5]
		fmt.Println("correct name!", n)
		walls = append(walls, n)
	}
	return walls, nil
}

// Note: name is already a PathEscaped string
func ReadWallet(name string) ([]byte, error) {
	return os.ReadFile("./" + name + ".keys")
}

// Note: name is already a PathEscaped string
func SaveWallet(name string, data []byte) error {
	return os.WriteFile("./"+name+".keys", data, 0o660)
}

func SaveRpcUrls(data []byte) error {
	return os.WriteFile("rpc-urls.txt", data, 0o660)
}
func ReadRpcUrls() ([]byte, error) {
	return os.ReadFile("rpc-urls.txt")
}
