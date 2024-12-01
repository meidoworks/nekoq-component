package secretapi

import (
	"testing"
)

func TestGenKeySet(t *testing.T) {
	keySet, err := DefaultKeyGen.GenerateVitalKeySet()
	if err != nil {
		t.Fatal(err)
	}
	if !keySet.VerifyCrc() {
		t.Fatal("Crc verification failed")
	}
	keySet.AES256[0] = keySet.AES256[0] + 1
	if keySet.VerifyCrc() {
		t.Fatal("Crc verification should fail")
	}
}
