package generalclient

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
)

type GetKeyResult struct {
	KeyId   string `json:"key_id"`
	KeyType string `json:"key_type"`
	Key     []byte `json:"key"`
	Format  string `json:"format"`
}

func (g *GetKeyResult) AesCipher() (cipher.Block, error) {
	if g.KeyType != "aes" {
		return nil, fmt.Errorf("unsupported key type: %s", g.KeyType)
	}
	return aes.NewCipher(g.Key)
}

func (g *GeneralClient) GetKeyByName(name string) (GetKeyResult, error) {
	req := SecretRequest{
		Method: "GET",
		Path:   "/api/v1/secret/key/name/" + name,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", g.opt.Secret.Token),
		},
	}

	resp, err := g.DoSecretJson(g.randomAddr(g.opt.Secret.AddrList), req)
	if err != nil {
		fmt.Println("Error:", err)
		return GetKeyResult{}, err
	}
	if resp.StatusCode != 200 {
		return GetKeyResult{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	r := GetKeyResult{}
	err = json.Unmarshal(resp.Body, &r)
	if err != nil {
		return GetKeyResult{}, err
	}
	return r, nil
}

func (g *GeneralClient) GetKeyById(id string) (GetKeyResult, error) {
	req := SecretRequest{
		Method: "GET",
		Path:   "/api/v1/secret/key/info/" + id,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", g.opt.Secret.Token),
		},
	}

	resp, err := g.DoSecretJson(g.randomAddr(g.opt.Secret.AddrList), req)
	if err != nil {
		fmt.Println("Error:", err)
		return GetKeyResult{}, err
	}
	if resp.StatusCode != 200 {
		return GetKeyResult{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	r := GetKeyResult{}
	err = json.Unmarshal(resp.Body, &r)
	if err != nil {
		return GetKeyResult{}, err
	}
	r.KeyId = id
	return r, nil
}
