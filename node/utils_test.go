package node

import (
	"bytes"
	. "github.com/franela/goblin"
	"math/big"
	"testing"
)

func TestUtils(t *testing.T) {
	g := Goblin(t)
	g.Describe("generateChordHash", func() {
		g.It("should generate M-bit hash", func() {
			s := "127.0.0.1:23333"
			h := generateChordHash(s)
			g.Assert(len(h)).Equal(MBytes)
		})
	})

	g.Describe("generateKardHash", func() {
		g.It("should generate 160-bit hash", func() {
			s := "127.0.0.1:23333"
			h := generateKadHash(s)
			g.Assert(len(h)).Equal(160 / 8)
		})
	})

	g.Describe("xorDistance", func() {
		g.It("should calculate distance", func() {
			a := make([]byte, 20)
			b := []byte("23333233332333323333")
			g.Assert(string(xorDistance(a, b))).Equal(string(b))
			a = []byte("23333233332333323333")
			b = []byte("23333233332333323333")
			g.Assert(bytes.Equal(xorDistance(a, b), []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})).IsTrue()
		})
	})

	g.Describe("inRange", func() {
		g.It("should not include left bound", func() {
			a := []byte("11111111111111111111")
			c := a
			b := []byte("22222222222222222222")
			g.Assert(inRange(c, a, b)).IsFalse()
		})

		g.It("should include right bound", func() {
			a := []byte("11111111111111111111")
			b := []byte("22222222222222222222")
			c := b
			g.Assert(inRange(c, a, b)).IsTrue()
		})

		g.It("should process cycle", func() {
			a := []byte("11111111111111111111")
			b := []byte("22222222222222222222")
			c := []byte("33333333333333333333")
			g.Assert(inRange(c, b, a)).IsTrue()
		})

		g.It("should include self in cycle", func() {
			a := []byte("11111111111111111111")
			b := []byte("11111111111111111111")
			c := []byte("11111111111111111111")
			g.Assert(inRange(c, a, b)).IsTrue()
			d := []byte("11111111121111111111")
			g.Assert(inRange(d, a, b)).IsTrue()
		})
	})

	g.Describe("inRangeExclude", func() {
		g.It("should include left bound", func() {
			a := []byte("11111111111111111111")
			c := a
			b := []byte("22222222222222222222")
			g.Assert(inRange(a, b, c)).IsTrue()
		})

		g.It("should include right bound", func() {
			a := []byte("11111111111111111111")
			b := []byte("22222222222222222222")
			c := b
			g.Assert(inRange(c, a, b)).IsTrue()
		})

		g.It("should process cycle", func() {
			a := []byte("11111111111111111111")
			b := []byte("22222222222222222222")
			c := []byte("33333333333333333333")
			g.Assert(inRange(c, b, a)).IsTrue()
		})

		g.It("should include everything in cycle", func() {
			a := []byte("11111111111111111111")
			b := []byte("11111111111111111111")
			c := []byte("11111111121111111111")
			g.Assert(inRangeExclude(c, a, b)).IsTrue()
		})
	})

	g.Describe("byteAddPowerOf2", func() {
		g.It("should add correctly", func() {
			a := big.NewInt(233)
			b := big.NewInt(0)
			b.SetBytes(byteAddPowerOf2(a.Bytes(), 10))
			g.Assert(b.Cmp(big.NewInt(233+1024)) == 0).IsTrue()
			b.SetBytes(byteAddPowerOf2(a.Bytes(), 16))
			if M > 16 {
				g.Assert(b.Cmp(big.NewInt(233+65536)) == 0).IsTrue()
			}
		})

		g.It("should add beyond bound", func() {
			a := big.NewInt(233)
			b := big.NewInt(0)
			b.SetBytes(byteAddPowerOf2(a.Bytes(), M+2))
			g.Assert(b.Cmp(big.NewInt(233)) == 0).IsTrue()
			a.Lsh(big.NewInt(1), M-1)
			b.SetBytes(byteAddPowerOf2(a.Bytes(), M-1))
			g.Assert(b.Cmp(big.NewInt(0)) == 0).IsTrue()
		})
	})
}
