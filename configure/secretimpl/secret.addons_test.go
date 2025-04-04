package secretimpl_test

import (
	"errors"
	"testing"
	"time"

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

	jwtToken, err := addonTool.SignJwtToken(jwtKey, secretapi.JwtAlgHS256, secretaddon.JwtClaims{
		"custome_data": "hello world!!!",
	}.SetID("id_111").SetIssuer("TestIssuer").SetSubject("TestSubject").SetAudience([]string{"TestAudience1", "TestAudience2"}).
		SetIat(time.Now()).SetExp(time.Now().Add(time.Hour)).SetNbf(time.Now()))
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

	jwtToken, err := addonTool.SignJwtToken(jwtKey, secretapi.JwtAlgHS512, map[string]interface{}{
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

	jwtToken, err := addonTool.SignJwtToken(jwtKey, secretapi.JwtAlgHS384, map[string]interface{}{
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
	algList := []secretaddon.JwtAlg{secretapi.JwtAlgRS256, secretapi.JwtAlgRS384, secretapi.JwtAlgRS512}

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
	algList := []secretaddon.JwtAlg{secretapi.JwtAlgPS256, secretapi.JwtAlgPS384, secretapi.JwtAlgPS512}

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
		{"test_addon_ecdsa256_key", secretapi.KeyECDSA256, secretapi.JwtAlgES256},
		{"test_addon_ecdsa384_key", secretapi.KeyECDSA384, secretapi.JwtAlgES384},
		{"test_addon_ecdsa521_key", secretapi.KeyECDSA521, secretapi.JwtAlgES512},
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
