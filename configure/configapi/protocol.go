package configapi

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

type Selectors struct {
	Data   map[string]string `cbor:"data,"`
	cached string
}

func (s *Selectors) cache() {
	size := len(s.Data)
	if size == 0 {
		return
	}
	m := s.Data
	if s.cached == "" {
		keySlice := make([]string, 0, size)
		for k, _ := range m {
			keySlice = append(keySlice, k)
		}
		slices.Sort(keySlice)
		buf := new(strings.Builder)
		buf.WriteString(keySlice[0])
		buf.WriteByte('=')
		buf.WriteString(m[keySlice[0]])
		for i := 1; i < size; i++ {
			buf.WriteByte(',')
			buf.WriteString(keySlice[i])
			buf.WriteByte('=')
			buf.WriteString(m[keySlice[i]])
		}
		s.cached = buf.String()
	}
}

type Configuration struct {
	// Group is a set of configurations
	Group string `cbor:"group,"`
	// Key is the name of the configuration
	Key string `cbor:"key,"`
	// Version represents the unique configuration history
	Version string `cbor:"version,"`

	// Value is the complete data of the configuration
	Value []byte `cbor:"value,"`
	// Signature represents the data integrity of the configuration
	Signature string `cbor:"sign,"`
	// Selector represents the part where the configuration will be used
	Selectors Selectors `cbor:"selectors,"`
	// OptionalSelectors is used for optional selectors matching
	OptionalSelectors Selectors `cbor:"opt_selectors,"`
	// Timestamp is the unix timestamp in second of the effective time(create/update) of this configuration
	Timestamp int64 `cbor:"timestamp,"`
}

func (c *Configuration) GenerateSignature() string {
	sha256sum := sha256.Sum256(c.Value)
	expectedSig := "sha256:" + hex.EncodeToString(sha256sum[:])
	return expectedSig
}

func (c *Configuration) ValidateSignature() bool {
	return c.Signature == c.GenerateSignature()
}

type RequestedConfigurationKey struct {
	Group   string `cbor:"group,"`
	Key     string `cbor:"key,"`
	Version string `cbor:"version,"`
}

type RawConfiguration struct {
	Group   string `cbor:"group,"`
	Key     string `cbor:"key,"`
	Version string `cbor:"version,"`
	Value   []byte `cbor:"value,"`
}
