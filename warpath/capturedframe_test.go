package warpath

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPack(t *testing.T) {
	expected := uint64(0x8765432187654321)
	b := []byte{
		0x87, 0x65, 0x43, 0x21,
		0x87, 0x65, 0x43, 0x21,
	}
	actual := pack(b)
	require.Equal(t, expected, actual)
}

func TestCalculateFSPLDistance(t *testing.T) {
	expected := 7.000397427391187 // m
	freq := 2412.0                // MHz
	sig := -57.0                  // dB
	actual := calculateFSPLDistance(sig, freq)
	require.Equal(t, expected, actual)
}
