package secretapi

import (
	"testing"
)

func TestCertSerialNumber(t *testing.T) {
	const N = 1001000000000
	var num CertSerialNumber
	num.FromInt64(N)
	i, err := num.ToBigInt()
	if err != nil {
		t.Fatal(err)
	}
	if i.Int64() != N {
		t.Fatal("serial number is not correct")
	}
	t.Log(num)

	num.FromBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19})
	t.Log(num)
	i, err = num.ToBigInt()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(i.String())
}
