package secret

import "testing"

func TestFormatAndScan(t *testing.T) {
	input := []byte{1, 2, 3, 4, 5, 6, 7, 8, 8, 7, 6, 5, 4, 3, 2, 1}
	inputNonce := []byte{9, 9, 9, 9, 9, 9, 9, 9}
	p := new(LocalFileUnsealProvider)
	str := p.formatEncryptedData(12, input, inputNonce)
	t.Logf(str)
	if str != "$12$01020304050607080807060504030201$0909090909090909" {
		t.Fatal("format encrypted data unexpected")
	}

	keyId, data, nonce, err := p.scanEncryptedData("$12$01020304050607080807060504030201$0909090909090909")
	if err != nil {
		t.Fatal(err)
	}
	if keyId != 12 {
		t.Fatal("key id unexpected")
	}
	for idx := range data {
		if data[idx] != input[idx] {
			t.Fatal("data unexpected")
		}
	}
	for idx := range nonce {
		if nonce[idx] != inputNonce[idx] {
			t.Fatal("nonce unexpected")
		}
	}
}
