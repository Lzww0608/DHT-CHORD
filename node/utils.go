package node

import (
	"bytes"
	"crypto/sha1"
	"math/big"
)

func generateChordHash(s string) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(s))
	bs := hasher.Sum(nil)
	return bs[:MBytes]
}

func generateKadHash(s string) []byte {
	hasher := sha1.New()
	hasher.Write([]byte(s))
	return hasher.Sum(nil)
}

func xorDistance(a, b []byte) []byte {
	c := make([]byte, len(a))
	for i := range a {
		c[i] = a[i] ^ b[i]
	}
	return c
}

func inRange(c, l, r []byte) bool {
	if bytes.Compare(l, r) < 0 {
		return bytes.Compare(l, c) < 0 && bytes.Compare(c, r) <= 0
	} else {
		return bytes.Compare(l, c) < 0 || bytes.Compare(c, r) <= 0
	}
}

func inRangeExclude(c, l, r []byte) bool {
	return !bytes.Equal(c, r) && inRange(c, l, r)
}

func byteAddPowerOf2(id []byte, exp uint) []byte {
	z := new(big.Int)
	z.SetBytes(id)
	p := new(big.Int)
	p.Lsh(big.NewInt(1), exp)
	z.Add(z, p)

	b := z.Bytes()
	result := make([]byte, MBytes)
	if len(b) > MBytes {
		for i := 0; i < MBytes; i++ {
			result[i] = b[i+len(b)-MBytes]
		}
	} else {
		i := 0
		for ; i < MBytes-len(b); i++ {
			result[i] = 0
		}
		for ; i < MBytes; i++ {
			result[i] = b[i+len(b)-MBytes]
		}
	}

	return result
}
