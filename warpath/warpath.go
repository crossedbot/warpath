package warpath

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

const (
	MAX_QUEUE_SIZE = 666
	MIN_QUEUE_SIZE = 10
)

// Warpath represents an interface to a warpath handler
type Warpath interface {
	Close()
	Run(filter string) (<-chan *CapturedFrame, error)
}

// warpath represents a warpath handler
type warpath struct {
	device string
	pcap_t *pcap.Handle
	quit   chan struct{}
}

// New returns a new Warpath handler, otherwise an error is returned.
func New(device string, snaplen int, timeout time.Duration) (Warpath, error) {
	var err error
	if device == "" {
		device, err = FindWirelessDevice()
		if err != nil {
			return nil, err
		}
	}
	if !canCapturePackets(device) {
		return nil, fmt.Errorf("cannot use device \"%s\" to capture packets", device)
	}
	pcap_t, err := activatePcap(device, snaplen, timeout)
	if err != nil {
		return nil, err
	}
	return &warpath{
		device: device,
		pcap_t: pcap_t,
		quit:   make(chan struct{}, MIN_QUEUE_SIZE),
	}, nil
}

// Close quit cpaturing, and close pcap handler
func (w *warpath) Close() {
	close(w.quit)
	w.pcap_t.Close()
}

// Run returns a channel of captured frames, otherwise an error is returned.
func (w *warpath) Run(filter string) (<-chan *CapturedFrame, error) {
	// create the output and packet source
	output := make(chan *CapturedFrame, MAX_QUEUE_SIZE)
	src, err := w.packetSource(filter)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case <-w.quit:
				// return if handler closed
				return
			case p := <-src.Packets():
				// parse the captured packets from source
				frame := NewCapturedFrame(&p)
				output <- frame
			}
		}
	}()
	return output, nil
}

// packetSource sets the BPF filter and returns a packet data source.
func (w *warpath) packetSource(filter string) (*gopacket.PacketSource, error) {
	// set the BPF filter
	if err := w.pcap_t.SetBPFFilter(filter); err != nil {
		return nil, err
	}
	// Use the handler as a packet source to process packets
	return gopacket.NewPacketSource(w.pcap_t, w.pcap_t.LinkType()), nil
}

// canCapturePackets checks if a device is suitable for capturing packets.
// Returns true if it can, otherwise false.
func canCapturePackets(device string) bool {
	return IsWirelessDevice(device) && CanSetRFMon(device)
}

// activatePcap returns an activated pcap handle, otherwise an error is returned.
func activatePcap(device string, snaplen int, timeout time.Duration) (*pcap.Handle, error) {
	// Create a new handler without activating it
	inactive, err := pcap.NewInactiveHandle(device)
	if err != nil {
		return nil, err
	}
	defer inactive.CleanUp()
	// Setup the device
	if err := inactive.SetSnapLen(snaplen); err != nil {
		return nil, err
	}
	if err := inactive.SetTimeout(timeout); err != nil {
		return nil, err
	}
	if err := inactive.SetRFMon(true); err != nil {
		return nil, err
	}
	if err := inactive.SetPromisc(true); err != nil {
		return nil, err
	}
	// Activate the device
	return inactive.Activate()
}
