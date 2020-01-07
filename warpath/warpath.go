package warpath

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type Warpath interface {
	Close()
	Start(filter string) error
	Pause()
	Resume()
	Stop()
	Output() <-chan *CapturedFrame
}

type warpath struct {
	device string
	pcap_t *pcap.Handle
	pause  chan struct{}
	resume chan struct{}
	quit   chan struct{}
	output chan *CapturedFrame
}

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
		pause:  make(chan struct{}, 10),
		resume: make(chan struct{}, 10),
		quit:   make(chan struct{}, 10),
		output: make(chan *CapturedFrame, 666),
	}, nil
}

func (w *warpath) Close() {
	close(w.pause)
	close(w.resume)
	close(w.quit)
	close(w.output)
	w.pcap_t.Close()
}

func (w *warpath) Start(filter string) error {
	src, err := w.packetSource(filter)
	if err != nil {
		return err
	}
	for {
		select {
		case <-w.quit:
			return nil
		case <-w.pause:
			<-w.resume
		case p := <-src.Packets():
			frame := NewCapturedFrame(&p)
			w.output <- frame
		}
	}
}

func (w *warpath) Pause() {
	w.pause <- struct{}{}
}

func (w *warpath) Resume() {
	w.resume <- struct{}{}
}

func (w *warpath) Stop() {
	w.Resume()
	w.quit <- struct{}{}
}

func (w *warpath) Output() <-chan *CapturedFrame {
	return w.output
}

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
