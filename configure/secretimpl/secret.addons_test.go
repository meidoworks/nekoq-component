package secretimpl_test

import (
	"errors"
	"testing"

	"github.com/meidoworks/nekoq-component/configure/secretaddon"
	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

func TestJwtCase1(t *testing.T) {

	keyStorage := InitTestKeyStorage(t)

	const jwtKey = "test_addon_64B_key"

	cipherTool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	if err := cipherTool.NewGeneral64BKey(jwtKey); err != nil {
		t.Fatal(err)
	}
	addonTool := secretaddon.NewAddonTool(keyStorage)

	jwtToken, err := addonTool.SignJwtToken(jwtKey, secretaddon.JwtAlgHS256, map[string]interface{}{
		"custome_data": "hello world!!!",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("jwt token:", jwtToken)

	claims, err := addonTool.VerifyJwtToken(jwtToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("claims:", claims)

	if err := cipherTool.NewGeneral64BKey(jwtKey); err != nil {
		t.Fatal(err)
	}

	claims, err = addonTool.VerifyJwtToken(jwtToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("claims:", claims)

}

func TestJwtCase2(t *testing.T) {

	keyStorage := InitTestKeyStorage(t)

	const jwtKey = "test_addon_128B_key"

	cipherTool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	if err := cipherTool.NewGeneral128BKey(jwtKey); err != nil {
		t.Fatal(err)
	}
	addonTool := secretaddon.NewAddonTool(keyStorage)

	jwtToken, err := addonTool.SignJwtToken(jwtKey, secretaddon.JwtAlgHS512, map[string]interface{}{
		"custome_data": "hello world!!!",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("jwt token:", jwtToken)

	claims, err := addonTool.VerifyJwtToken(jwtToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("claims:", claims)

	if err := cipherTool.NewGeneral128BKey(jwtKey); err != nil {
		t.Fatal(err)
	}

	claims, err = addonTool.VerifyJwtToken(jwtToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("claims:", claims)

}

func TestJwtCase3(t *testing.T) {

	keyStorage := InitTestKeyStorage(t)

	const jwtKey = "test_addon_128B_key"

	cipherTool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	if err := cipherTool.NewGeneral128BKey(jwtKey); err != nil {
		t.Fatal(err)
	}
	addonTool := secretaddon.NewAddonTool(keyStorage)

	jwtToken, err := addonTool.SignJwtToken(jwtKey, secretaddon.JwtAlgHS384, map[string]interface{}{
		"custome_data": "hello world!!!",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("jwt token:", jwtToken)

	claims, err := addonTool.VerifyJwtToken(jwtToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("claims:", claims)

	if err := cipherTool.NewGeneral128BKey(jwtKey); err != nil {
		t.Fatal(err)
	}

	claims, err = addonTool.VerifyJwtToken(jwtToken)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("claims:", claims)

}

func TestJwtCase4(t *testing.T) {
	keyStorage := InitTestKeyStorage(t)
	keys := []struct {
		Key     string
		KeyType secretapi.KeyType
	}{
		{"test_addon_rsa1024_key", secretapi.KeyRSA1024},
		{"test_addon_rsa2048_key", secretapi.KeyRSA2048},
		{"test_addon_rsa3072_key", secretapi.KeyRSA3072},
		{"test_addon_rsa4096_key", secretapi.KeyRSA4096},
	}
	algList := []secretaddon.JwtAlg{secretaddon.JwtAlgRS256, secretaddon.JwtAlgRS384, secretaddon.JwtAlgRS512}

	cipherTool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	addonTool := secretaddon.NewAddonTool(keyStorage)

	f := func(t *testing.T, testCase struct {
		Key     string
		KeyType secretapi.KeyType
	}, alg secretaddon.JwtAlg) {
		t.Log("===========>>>>> ", testCase.Key, alg)

		if err := cipherTool.NewRsaKey(testCase.Key, testCase.KeyType); err != nil {
			t.Fatal(err)
		}

		jwtToken, err := addonTool.SignJwtToken(testCase.Key, alg, map[string]interface{}{
			"custome_data": "hello world!!!",
		})
		if err != nil {
			t.Fatal(err)
		}
		t.Log("jwt token:", jwtToken)

		claims, err := addonTool.VerifyJwtToken(jwtToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("claims:", claims)

		if err := cipherTool.NewRsaKey(testCase.Key, testCase.KeyType); err != nil {
			t.Fatal(err)
		}

		claims, err = addonTool.VerifyJwtToken(jwtToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("claims:", claims)
	}

	for _, testCase := range keys {
		for _, alg := range algList {
			f(t, testCase, alg)
		}
	}
}

func TestJwtCase5(t *testing.T) {
	keyStorage := InitTestKeyStorage(t)
	keys := []struct {
		Key     string
		KeyType secretapi.KeyType
	}{
		{"test_addon_rsa1024_key", secretapi.KeyRSA1024},
		{"test_addon_rsa2048_key", secretapi.KeyRSA2048},
		{"test_addon_rsa3072_key", secretapi.KeyRSA3072},
		{"test_addon_rsa4096_key", secretapi.KeyRSA4096},
	}
	algList := []secretaddon.JwtAlg{secretaddon.JwtAlgPS256, secretaddon.JwtAlgPS384, secretaddon.JwtAlgPS512}

	cipherTool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	addonTool := secretaddon.NewAddonTool(keyStorage)

	f := func(t *testing.T, testCase struct {
		Key     string
		KeyType secretapi.KeyType
	}, alg secretaddon.JwtAlg) {
		t.Log("===========>>>>> ", testCase.Key, alg)

		if err := cipherTool.NewRsaKey(testCase.Key, testCase.KeyType); err != nil {
			t.Fatal(err)
		}

		jwtToken, err := addonTool.SignJwtToken(testCase.Key, alg, map[string]interface{}{
			"custome_data": "hello world!!!",
		})
		if errors.Is(err, secretaddon.ErrJwtInvalidAlgCombination) {
			t.Log("invalid jwt alg combination, skip")
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Log("jwt token:", jwtToken)

		claims, err := addonTool.VerifyJwtToken(jwtToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("claims:", claims)

		if err := cipherTool.NewRsaKey(testCase.Key, testCase.KeyType); err != nil {
			t.Fatal(err)
		}

		claims, err = addonTool.VerifyJwtToken(jwtToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("claims:", claims)
	}

	for _, testCase := range keys {
		for _, alg := range algList {
			f(t, testCase, alg)
		}
	}
}

func TestJwtCase6(t *testing.T) {
	keyStorage := InitTestKeyStorage(t)
	keys := []struct {
		Key     string
		KeyType secretapi.KeyType
		JwtAlg  secretaddon.JwtAlg
	}{
		{"test_addon_ecdsa256_key", secretapi.KeyECDSA256, secretaddon.JwtAlgES256},
		{"test_addon_ecdsa384_key", secretapi.KeyECDSA384, secretaddon.JwtAlgES384},
		{"test_addon_ecdsa521_key", secretapi.KeyECDSA521, secretaddon.JwtAlgES512},
	}

	cipherTool := secretapi.NewLevel2CipherTool(keyStorage, secretapi.DefaultKeyGen, "test_case")
	addonTool := secretaddon.NewAddonTool(keyStorage)

	f := func(t *testing.T, testCase struct {
		Key     string
		KeyType secretapi.KeyType
		JwtAlg  secretaddon.JwtAlg
	}) {
		t.Log("===========>>>>> ", testCase.Key, testCase.JwtAlg)

		if err := cipherTool.NewEcdsaKey(testCase.Key, testCase.KeyType); err != nil {
			t.Fatal(err)
		}

		jwtToken, err := addonTool.SignJwtToken(testCase.Key, testCase.JwtAlg, map[string]interface{}{
			"custome_data": "hello world!!!",
		})
		if errors.Is(err, secretaddon.ErrJwtInvalidAlgCombination) {
			t.Log("invalid jwt alg combination, skip")
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		t.Log("jwt token:", jwtToken)

		claims, err := addonTool.VerifyJwtToken(jwtToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("claims:", claims)

		if err := cipherTool.NewEcdsaKey(testCase.Key, testCase.KeyType); err != nil {
			t.Fatal(err)
		}

		claims, err = addonTool.VerifyJwtToken(jwtToken)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("claims:", claims)
	}

	for _, testCase := range keys {
		f(t, testCase)
	}
}
