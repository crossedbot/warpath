package warpath

import (
	"net"
	"sync"
)

type HwAddrMap struct {
	sync.RWMutex
	m map[uint64]net.HardwareAddr
}

// NewHwAddrMap returns a newly instantiated HwAddrMap.
func NewHwAddrMap() *HwAddrMap {
	return &HwAddrMap{m: make(map[uint64]net.HardwareAddr)}
}

// Del deletes the hardware address from the map.
func (hwMap *HwAddrMap) Del(hwaddr net.HardwareAddr) {
	hash := HashHwAddr(hwaddr)
	hwMap.Lock()
	delete(hwMap.m, hash)
	hwMap.Unlock()
}

// Get returns the key and value for a given hardware address in the hardware
// address map.
func (hwMap *HwAddrMap) Get(hwaddr net.HardwareAddr) (uint64, net.HardwareAddr, bool) {
	var ok bool = false
	hash := HashHwAddr(hwaddr)
	hwMap.RLock()
	addr := hwMap.m[hash]
	hwMap.RUnlock()
	if addr != nil {
		ok = true
	}
	return hash, addr, ok
}

// Add inserts the hardware address as an item into the hardware address map.
// Returns the new hardware address map that includes the newly inserted item.
func (hwMap *HwAddrMap) Set(hwaddr net.HardwareAddr) {
	hash := HashHwAddr(hwaddr)
	hwMap.Lock()
	hwMap.m[hash] = hwaddr
	hwMap.Unlock()
}

// HashHwAddr returns a hashed hardware address.
func HashHwAddr(hwaddr net.HardwareAddr) uint64 {
	var hash, oui, nic uint64
	oui = uint64(hwaddr[0])<<16 | uint64(hwaddr[1])<<8 | uint64(hwaddr[2])
	nic = uint64(hwaddr[3])<<16 | uint64(hwaddr[4])<<8 | uint64(hwaddr[5])
	hash = (oui ^ nic)
	return hash
}

// GetHwAddr returns tehe hardware address for a given device by name, otherwise
// an error is returned.
func GetHwAddr(deviceName string) (net.HardwareAddr, error) {
	device, err := net.InterfaceByName(deviceName)
	if err != nil {
		return net.HardwareAddr{}, err
	}
	return device.HardwareAddr, nil
}
