package generalclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type SecretRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
}

type SecretResponse struct {
	StatusCode int
	Body       []byte
}

func (c *GeneralClient) DoSecretJson(baseUrl string, req SecretRequest) (*SecretResponse, error) {
	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequest(req.Method, baseUrl+req.Path, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Perform the HTTP request
	resp, err := c.secretHttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &SecretResponse{
		StatusCode: resp.StatusCode,
		Body:       responseBody,
	}, nil
}

type NewCertificateRequest struct {
	KeyName string `json:"key_name"` // the key should be L2 key

	CommonName     string   `json:"cn"`
	Organization   string   `json:"org"`
	Country        string   `json:"country"`
	Province       string   `json:"province"`
	Locality       string   `json:"locality"`
	StreetAddress  string   `json:"street_address"`
	PostalCode     string   `json:"postal_code"`
	DNSNames       []string `json:"dns_names"`
	IPAddresses    []string `json:"ip_addresses"`
	EmailAddresses []string `json:"email_addresses"`
}

type NewCertificateRequestResult struct {
	KeyId   string `json:"key_id"`
	Request string `json:"req"`
}

func (g *GeneralClient) NewCertReq(model NewCertificateRequest) (NewCertificateRequestResult, error) {
	req := SecretRequest{
		Method: "POST",
		Path:   "/api/v1/secret/cert/newreq",
		Headers: map[string]string{
			"Accept":        "application/json",
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", g.opt.Secret.Token),
		},
		Body: model,
	}

	resp, err := g.DoSecretJson(g.randomAddr(g.opt.Secret.AddrList), req)
	if err != nil {
		fmt.Println("Error:", err)
		return NewCertificateRequestResult{}, err
	}
	if resp.StatusCode != 200 {
		return NewCertificateRequestResult{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	r := NewCertificateRequestResult{}
	err = json.Unmarshal(resp.Body, &r)
	if err != nil {
		return NewCertificateRequestResult{}, err
	}
	return r, nil
}

type NewCertificate struct {
	CAName      string `json:"ca_name"` // ca cert should have level2 key
	TTL         int    `json:"ttl"`
	CertName    string `json:"cert_name"`
	CertReqData string `json:"cert_req_data"`
	CertUsage   string `json:"cert_usage"` // available: server/client/both

	// support managed private key
	KeyId string `json:"key_id"`
}

type NewCertificateResult struct {
	Cert   string   `json:"cert"`
	CaList []string `json:"ca_list"`
}

func (g *GeneralClient) NewCert(model NewCertificate) (NewCertificateResult, error) {
	req := SecretRequest{
		Method: "POST",
		Path:   "/api/v1/secret/cert/new",
		Headers: map[string]string{
			"Accept":        "application/json",
			"Content-Type":  "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", g.opt.Secret.Token),
		},
		Body: model,
	}

	resp, err := g.DoSecretJson(g.randomAddr(g.opt.Secret.AddrList), req)
	if err != nil {
		fmt.Println("Error:", err)
		return NewCertificateResult{}, err
	}
	if resp.StatusCode != 200 {
		return NewCertificateResult{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	r := NewCertificateResult{}
	err = json.Unmarshal(resp.Body, &r)
	if err != nil {
		return NewCertificateResult{}, err
	}
	return r, nil
}

type GetCertificateResult struct {
	Cert   string   `json:"cert"`
	CaList []string `json:"ca_list"`
	Key    string   `json:"key"`
}

func (g *GeneralClient) GetCert(certName string) (GetCertificateResult, error) {
	req := SecretRequest{
		Method: "GET",
		Path:   "/api/v1/secret/cert/name/" + certName,
		Headers: map[string]string{
			"Accept":        "application/json",
			"Authorization": fmt.Sprintf("Bearer %s", g.opt.Secret.Token),
		},
	}

	resp, err := g.DoSecretJson(g.randomAddr(g.opt.Secret.AddrList), req)
	if err != nil {
		fmt.Println("Error:", err)
		return GetCertificateResult{}, err
	}
	if resp.StatusCode != 200 {
		return GetCertificateResult{}, fmt.Errorf("request failed with status code %d", resp.StatusCode)
	}

	r := GetCertificateResult{}
	err = json.Unmarshal(resp.Body, &r)
	if err != nil {
		return GetCertificateResult{}, err
	}
	return r, nil
}
