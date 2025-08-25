//go:build js
// +build js

package save

import (
	"encoding/base64"
	"fmt"
	"syscall/js"
)

func GetWallets() ([]string, error) {
	// Get all keys from localStorage
	localStorage := js.Global().Get("localStorage")
	length := localStorage.Get("length").Int()

	var wallets []string

	for i := 0; i < length; i++ {
		key := localStorage.Call("key", i).String()
		// Check if key ends with .keys
		if len(key) >= 5 && key[len(key)-5:] == ".keys" {
			// Remove .keys suffix
			name := key[:len(key)-5]
			wallets = append(wallets, name)
		}
	}

	return wallets, nil
}

// Note: name is already a PathEscaped string
func ReadWallet(name string) ([]byte, error) {
	localStorage := js.Global().Get("localStorage")
	item := localStorage.Call("getItem", name+".keys")

	if item.IsNull() {
		return nil, fmt.Errorf("wallet not found: %s", name)
	}

	data := item.String()
	return base64.URLEncoding.DecodeString(data)
}

// Note: name is already a PathEscaped string
func SaveWallet(name string, data []byte) error {
	localStorage := js.Global().Get("localStorage")
	dataStr := base64.URLEncoding.EncodeToString(data)
	localStorage.Call("setItem", name+".keys", dataStr)
	return nil
}

func SaveRpcUrls(data []byte) error {
	localStorage := js.Global().Get("localStorage")
	dataStr := base64.URLEncoding.EncodeToString(data)
	localStorage.Call("setItem", "rpc-urls.txt", dataStr)
	return nil
}

func ReadRpcUrls() ([]byte, error) {
	localStorage := js.Global().Get("localStorage")
	item := localStorage.Call("getItem", "rpc-urls.txt")

	if item.IsNull() {
		return nil, fmt.Errorf("rpc-urls.txt not found")
	}

	data := item.String()
	return base64.URLEncoding.DecodeString(data)
}
