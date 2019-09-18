package imagemanip

import (
	"math/rand"
	"testing"
)

func TestInvert(t *testing.T) {
	for i := uint64(1); i < modulus; i++ {
		if i*invert(i)%modulus != 1 {
			t.Fatalf("Failed to invert %v", i)
		}
	}
}

func TestFit1(t *testing.T) {
	invec := make([]uint16, 4096)
	tvec := make([]uint16, 4096)
	for i := 0; i < 4096; i++ {
		invec[i] = uint16(rand.Intn(int(modulus)))
		tvec[i] = invec[i]
	}
	fit1(tvec, 0, 1, 12)
	iit1(tvec, 0, 1, 12)

	for i := 0; i < 4096; i++ {
		if invec[i] != tvec[i] {
			t.Fatalf("Mismatch at position %v", i)
		}
	}
}

func TestFit2(t *testing.T) {
	invec := make([]uint16, 512*512)
	tvec := make([]uint16, 512*512)
	for idx, _ := range(invec) {
		invec[idx] = uint16(rand.Intn(int(modulus)))
		tvec[idx] = invec[idx]
	}

	mi := modulusImage{
		pixels: invec,
		width:  512,
	}
	fit2(mi)
	iit2(mi)
	for idx, _ := range mi.pixels {
		if mi.pixels[idx] != tvec[idx] {
			t.Fatalf("Mismatch at position %v", idx)
		}
	}
}