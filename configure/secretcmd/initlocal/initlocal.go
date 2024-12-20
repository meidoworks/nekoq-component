package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

var (
	inputFile  string
	outputFile string
)

func init() {
	flag.StringVar(&inputFile, "input", "", "input file")
	flag.StringVar(&outputFile, "output", "", "output file")
	flag.Parse()
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
	data, err := new(secretapi.PemTool).EncodeKeySet(ks)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
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

	ks, err := new(secretapi.PemTool).ParseKeySet(data)
	if err != nil {
		panic(err)
	}
	keySetData, err := ks.SaveAsBytes()
	if err != nil {
		panic(err)
	}

	output, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	defer func(output *os.File) {
		_ = output.Close()
	}(output)
	_, err = output.Write(keySetData)
	if err != nil {
		panic(err)
	}
	if err := output.Sync(); err != nil {
		panic(err)
	}
}
