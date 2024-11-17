package configclient

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

var (
	ErrNoSuchEnvironmentVariable                = errors.New("no such environment variable")
	ErrHasDuplicatedKeysWithDifferentCaseFormat = errors.New("duplicated keys with different case format")
)

type EnvClient struct {
	envs              map[string]string
	inputKeyTransform func(string) string

	parserWarning error
}

func NewEnvClient() *EnvClient {
	list := os.Environ()

	var envs = make(map[string]string)
	for _, v := range list {
		kv := strings.SplitN(v, "=", 2)
		if len(kv) != 2 {
			panic("invalid environment variable:" + v)
		}
		envs[kv[0]] = kv[1]
	}

	ec := &EnvClient{
		envs: envs,
		inputKeyTransform: func(key string) string {
			return key
		},
	}
	return ec
}

func (e *EnvClient) ParserWarning() error {
	return e.parserWarning
}

func (c *EnvClient) GetString(key string) (string, error) {
	value, ok := c.envs[c.inputKeyTransform(key)]
	if !ok {
		return "", ErrNoSuchEnvironmentVariable
	}
	return value, nil
}

func (c *EnvClient) GetInt(key string) (int, error) {
	value, ok := c.envs[c.inputKeyTransform(key)]
	if !ok {
		return 0, ErrNoSuchEnvironmentVariable
	}
	return strconv.Atoi(value)
}

func (c *EnvClient) GetBool(key string) (bool, error) {
	value, ok := c.envs[c.inputKeyTransform(key)]
	if !ok {
		return false, ErrNoSuchEnvironmentVariable
	}
	return strconv.ParseBool(value)
}

func (c *EnvClient) GetFloat(key string) (float64, error) {
	value, ok := c.envs[c.inputKeyTransform(key)]
	if !ok {
		return 0, ErrNoSuchEnvironmentVariable
	}
	return strconv.ParseFloat(value, 64)
}

func (c *EnvClient) GetInt64(key string) (int64, error) {
	value, ok := c.envs[c.inputKeyTransform(key)]
	if !ok {
		return 0, ErrNoSuchEnvironmentVariable
	}
	return strconv.ParseInt(value, 10, 64)
}

func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}
	return val
}

type defaultVal[T any] struct {
	val T
}

func (d defaultVal[T]) On(val T, err error) T {
	if errors.Is(err, ErrNoSuchEnvironmentVariable) {
		return d.val
	} else if err != nil {
		panic(err)
	} else {
		return val
	}
}

func MustDefault[T any](dv T) defaultVal[T] {
	return defaultVal[T]{val: dv}
}

func (c *EnvClient) CaseInsensitive() *EnvClient {
	trans := func(key string) string {
		return strings.ToLower(key)
	}
	var envs = make(map[string]string)
	for k, v := range c.envs {
		envs[trans(k)] = v
	}
	var parserWarning error
	if len(c.envs) != len(envs) {
		parserWarning = ErrHasDuplicatedKeysWithDifferentCaseFormat
	}
	return &EnvClient{
		envs:              envs,
		inputKeyTransform: trans,
		parserWarning:     parserWarning,
	}
}
