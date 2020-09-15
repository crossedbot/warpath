package warpath

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHwAddrMapGet(t *testing.T) {
	hwMap := NewHwAddrMap()
	hwaddr := net.HardwareAddr{0x11, 0x22, 0xDE, 0xAD, 0xBE, 0xEF}
	hash := uint64(0x1122DE ^ 0xADBEEF)
	hwMap.m[hash] = hwaddr
	key, value, isSet := hwMap.Get(hwaddr)
	require.True(t, isSet)
	require.Equal(t, hash, key)
	require.Equal(t, hwaddr, value)
}

func TestHwAddrMapSet(t *testing.T) {
	hwMap := NewHwAddrMap()
	hwaddr := net.HardwareAddr{0x11, 0x22, 0xDE, 0xAD, 0xBE, 0xEF}
	hash := uint64(0x1122DE ^ 0xADBEEF)
	hwMap.Set(hwaddr)
	require.Equal(t, hwaddr, hwMap.m[hash])
}

func TestHwAddrMapDel(t *testing.T) {
	hwMap := NewHwAddrMap()
	hwaddr := net.HardwareAddr{0x11, 0x22, 0xDE, 0xAD, 0xBE, 0xEF}
	hash := uint64(0x1122DE ^ 0xADBEEF)
	hwMap.Set(hwaddr)
	require.Equal(t, hwaddr, hwMap.m[hash])
	hwMap.Del(hwaddr)
	require.Nil(t, hwMap.m[hash])
}

func TestHashHwAddr(t *testing.T) {
	hwaddr := net.HardwareAddr{0x11, 0x22, 0xDE, 0xAD, 0xBE, 0xEF}
	hash := uint64(0x1122DE ^ 0xADBEEF)
	require.Equal(t, hash, HashHwAddr(hwaddr))
}
