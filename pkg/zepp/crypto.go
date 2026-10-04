package zepp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

// Zepp Life 6.12.0's default network key (version 1), decoded by
// Crypto.method02 from Nqvk+0fJ1bunqWW9X7FWhw== in HmKeyManager.
// The native method03 uses AES-128-CBC with PKCS#7 padding and raw bytes.
const accountKey = "xeNtBVqzDc6tuNTh"
const accountIV = "MAAAYAAAAAAAAABg"

func encryptAccountForm(form string) []byte {
	block, _ := aes.NewCipher([]byte(accountKey))
	data := []byte(form)
	padding := aes.BlockSize - len(data)%aes.BlockSize
	data = append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)
	cipher.NewCBCEncrypter(block, []byte(accountIV)).CryptBlocks(data, data)
	return data
}
