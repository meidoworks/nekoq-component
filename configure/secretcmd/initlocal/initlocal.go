package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/fxamacker/cbor/v2"
	"golang.org/x/crypto/pbkdf2"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

type EncBlock struct {
	S []byte `cbor:"S"`
	T []byte `cbor:"T"`
}

var (
	inputFile  string
	outputFile string
	password   string
)

func init() {
	flag.StringVar(&inputFile, "input", "", "input file")
	flag.StringVar(&outputFile, "output", "", "output file")
	flag.StringVar(&password, "password", "", "password")
	flag.Parse()
}

func keyGen(password string) ([]byte, []byte) {
	salt := make([]byte, 128)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	return pbkdf2.Key([]byte(password), salt, 4096, 32, sha256.New), salt
}

func encryptData(data []byte, password string) []byte {
	key, salt := keyGen(password)
	ci, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aead, err := cipher.NewGCM(ci)
	if err != nil {
		panic(err)
	}
	encrypted := aead.Seal(nil, salt[:aead.NonceSize()], data, nil)

	block := &EncBlock{salt, encrypted}
	result, err := cbor.Marshal(block)
	if err != nil {
		panic(err)
	}
	return result
}

func decryptData(result []byte, password string) []byte {
	block := &EncBlock{}
	if err := cbor.Unmarshal(result, block); err != nil {
		panic(err)
	}
	key := pbkdf2.Key([]byte(password), block.S, 4096, 32, sha256.New)
	ci, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aead, err := cipher.NewGCM(ci)
	if err != nil {
		panic(err)
	}
	plaintext, err := aead.Open(nil, block.S[:aead.NonceSize()], block.T, nil)
	if err != nil {
		panic(err)
	}
	return plaintext
}

// This tool is used for the two purpose:
// 1. Generate new KeySet in the format of PEM
// 2. Convert the PEM KeySet file into available file for Unseal Local Provider
func main() {
	if inputFile != "" || outputFile != "" {
		decodeMode()
		return
	}
	ks, err := secretapi.DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		panic(err)
	}
	data, err := ks.SaveAsBytes()
	if err != nil {
		panic(err)
	}
	encData := encryptData(data, password)

	result, err := new(secretapi.PemTool).EncodeCustom(encData, "ENCRYPTED KETSET")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(result))
}

func decodeMode() {
	input, err := os.Open(inputFile)
	if err != nil {
		panic(err)
	}
	defer func(input *os.File) {
		_ = input.Close()
	}(input)
	data, err := io.ReadAll(input)
	if err != nil {
		panic(err)
	}

	encData, err := new(secretapi.PemTool).ParseCustom(data)
	if err != nil {
		panic(err)
	}
	plaintext := decryptData(encData, password)

	output, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer func(output *os.File) {
		_ = output.Close()
	}(output)
	_, err = output.Write(plaintext)
	if err != nil {
		panic(err)
	}
	if err := output.Sync(); err != nil {
		panic(err)
	}
}
