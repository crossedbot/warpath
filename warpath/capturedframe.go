package warpath

import (
	"math"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// CapturedFrame represents a captured frame from a Wifi device
type CapturedFrame struct {
	ID        int64   `json:"-"`
	Type      string  `json:"type"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance"`
	Timestamp int64   `json:"timestamp"`
	Source    uint64  `json:"source"`
	Data      []byte  `json:"data"`
}

// NewCapturedFrame parses a gopacket and returns a captured frame.
func NewCapturedFrame(packet *gopacket.Packet) (frame *CapturedFrame) {
	frame = new(CapturedFrame)
	rt := (*packet).Layer(layers.LayerTypeRadioTap)
	if rt != nil {
		// if a radiotap frame exists, parse it
		if rt, ok := rt.(*layers.RadioTap); ok {
			if rt.Present.DBMAntennaSignal() && rt.Present.Channel() {
				dbAntSignal := int(rt.DBAntennaSignal)
				if dbAntSignal >= 127 {
					dbAntSignal = dbAntSignal - 255
				}
				// estimate the distance from the wifi device using the
				// Free-space path loss
				frame.Distance = calculateFSPLDistance(
					float64(dbAntSignal),
					float64(rt.ChannelFrequency),
				)
			}
		}
	}
	// parse the 802.11 frame
	dot11 := (*packet).Layer(layers.LayerTypeDot11)
	if dot11 != nil {
		if dot11, ok := dot11.(*layers.Dot11); ok {
			frame.Type = dot11.Type.String()
			// Address 2 = Source Address; Address 3 = BSSID
			frame.Source = pack([]byte(dot11.Address2))
		}
	}
	frame.Timestamp = time.Now().Unix()
	frame.Data = (*packet).Data()
	return
}

// pack packs a byte array as a uint64 value.
func pack(b []byte) uint64 {
	var p uint64
	if len(b) > 0 && len(b) <= 8 {
		p = uint64(b[0])
		for _, v := range b[1:] {
			p = (p << 8) | uint64(v)
		}
	}
	return p
}

// calculateFSPLDistance returns the calculated distance (in meters) using the
// FSPL formula for a device using the antenna signal (dB) and frequency (MHz).
func calculateFSPLDistance(antsignal, freq float64) float64 {
	exp := ((27.55 - (20 * math.Log10(freq))) + math.Abs(antsignal)) / 20.0
	return math.Pow(10.0, exp)
}
