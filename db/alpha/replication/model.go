package replication

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"math"
)

type Position [2]int64

func (p Position) ToBytes() []byte {
	r := make([]byte, 16)
	binary.BigEndian.PutUint64(r[0:8], uint64(p[0]&math.MaxInt64))
	binary.BigEndian.PutUint64(r[8:16], uint64(p[1]&math.MaxInt64))
	return r
}

func (p *Position) FromBytes(data []byte) error {
	if len(data) != 16 {
		return errors.New("invalid data length")
	}
	p[0] = int64(binary.BigEndian.Uint64(data[0:8]) & math.MaxInt64)
	p[1] = int64(binary.BigEndian.Uint64(data[8:16]) & math.MaxInt64)
	return nil
}

func (p Position) FirstToHexString() string {
	return hex.EncodeToString(p.ToBytes()[:8])
}

func (p Position) LastToHexString() string {
	return hex.EncodeToString(p.ToBytes()[8:])
}
