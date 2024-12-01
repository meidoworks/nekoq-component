package secret

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/meidoworks/nekoq-component/configure/secretapi"
)

type LocalFileUnsealProvider struct {
	keySets     map[int64]*secretapi.KeySet
	latestKeyId int64
	latestKey   *secretapi.KeySet
}

func NewLocalFileUnsealProvider(fsfs fs.FS, keyFileMapping map[int64]string) (*LocalFileUnsealProvider, error) {
	var keySets = map[int64]*secretapi.KeySet{}
	for keyId, filePath := range keyFileMapping {
		fn := func() (result []byte, rerr error) {
			f, err := fsfs.Open(filePath)
			if err != nil {
				return nil, err
			}
			defer func(f fs.File) {
				err := f.Close()
				if err != nil {
					rerr = err
				}
			}(f)
			data, err := io.ReadAll(f)
			if err != nil {
				return nil, err
			}
			return data, nil
		}
		data, err := fn()
		if err != nil {
			return nil, err
		}
		keySet := new(secretapi.KeySet)
		if err := keySet.LoadFromBytes(data); err != nil {
			return nil, err
		}
		if !keySet.VerifyCrc() {
			return nil, fmt.Errorf("keySet does not have the expected crc")
		}
		keySets[keyId] = keySet
	}

	keys := maps.Keys(keySets)
	maxKeyId := slices.Max(slices.Collect(keys))

	return &LocalFileUnsealProvider{
		keySets:     keySets,
		latestKey:   keySets[maxKeyId],
		latestKeyId: maxKeyId,
	}, nil
}

func (l *LocalFileUnsealProvider) WaitUnsealOperation(ctx context.Context, encToken string) (*secretapi.UnsealResponse, error) {
	if encToken == "" {
		// initial state
		return new(secretapi.UnsealResponse), nil
	}

	if len(l.keySets) == 0 {
		return nil, fmt.Errorf("no secret unseal")
	}
	plaintext, keyId, err := l.decryptInternal(ctx, []byte(encToken))
	if err != nil {
		return nil, err
	}
	return &secretapi.UnsealResponse{
		Selector: strconv.Itoa(int(keyId)),
		Token:    string(plaintext),
	}, nil
}

func (l *LocalFileUnsealProvider) Encrypt(ctx context.Context, plaintext []byte) ([]byte, string, error) {
	encData, nonce, err := l.latestKey.AesGCMEnc(plaintext)
	if err != nil {
		return nil, "", err
	}
	return []byte(l.formatEncryptedData(l.latestKeyId, encData, nonce)), l.UseKeyId(), nil
}

func (l *LocalFileUnsealProvider) decryptInternal(ctx context.Context, ciphertext []byte) ([]byte, int64, error) {
	var ciphertextString = string(ciphertext)
	keyId, data, nonce, err := l.scanEncryptedData(ciphertextString)
	if err != nil {
		return nil, 0, err
	}
	keyset, ok := l.keySets[keyId]
	if !ok {
		return nil, 0, fmt.Errorf("no recent KeySet provided for decrypting token with keyId %d", keyId)
	}
	result, err := keyset.AesGCMDec(data, nonce)
	return result, keyId, err
}

func (l *LocalFileUnsealProvider) Decrypt(ctx context.Context, ciphertext []byte) ([]byte, error) {
	result, _, err := l.decryptInternal(ctx, ciphertext)
	return result, err
}

func (l *LocalFileUnsealProvider) formatEncryptedData(keyId int64, data []byte, nonce []byte) string {
	return fmt.Sprintf("$%d$%s$%s", keyId, hex.EncodeToString(data), hex.EncodeToString(nonce))
}

func (l *LocalFileUnsealProvider) scanEncryptedData(str string) (int64, []byte, []byte, error) {
	var keyId int64
	var dataStr string
	if _, err := fmt.Sscanf(str, "$%d$%s", &keyId, &dataStr); err != nil {
		return 0, nil, nil, err
	}
	splits := strings.Split(dataStr, "$")
	if len(splits) != 2 {
		return 0, nil, nil, fmt.Errorf("invalid encrypted data and nonce")
	}
	data, err := hex.DecodeString(splits[0])
	if err != nil {
		return 0, nil, nil, err
	}
	nonce, err := hex.DecodeString(splits[1])
	if err != nil {
		return 0, nil, nil, err
	}
	return keyId, data, nonce, nil
}

func (l *LocalFileUnsealProvider) UseKeyId() string {
	return fmt.Sprint(l.latestKeyId)
}
