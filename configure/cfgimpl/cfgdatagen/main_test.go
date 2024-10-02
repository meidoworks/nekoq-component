package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/fxamacker/cbor/v2"

	"github.com/meidoworks/nekoq-component/configure/configapi"
)

var testCfg = &configapi.Configuration{
	Group:     "group_" + fmt.Sprint(rand.Int()),
	Key:       "key_" + fmt.Sprint(rand.Int()),
	Version:   "v1.1",
	Value:     []byte("test data"),
	Signature: "testsig1",
	Selectors: configapi.Selectors{
		Data: map[string]string{
			"dc": "dc1",
		},
	},
	OptionalSelectors: configapi.Selectors{},
	Timestamp:         time.Now().Unix(),
}

func BenchmarkJsonMarshalling(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(testCfg)
	}
}

func BenchmarkCborMarshalling(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = cbor.Marshal(testCfg)
	}
}

func BenchmarkJsonUnmarshalling(b *testing.B) {
	data, err := json.Marshal(testCfg)
	if err != nil {
		b.Fatal(err)
	}
	cfg := new(configapi.Configuration)
	for i := 0; i < b.N; i++ {
		_ = json.Unmarshal(data, cfg)
	}
}

func BenchmarkCborUnmarshalling(b *testing.B) {
	data, err := cbor.Marshal(testCfg)
	if err != nil {
		b.Fatal(err)
	}
	cfg := new(configapi.Configuration)
	for i := 0; i < b.N; i++ {
		_ = cbor.Unmarshal(data, cfg)
	}
}
