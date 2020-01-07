package warpath

import (
	"errors"

	"github.com/google/gopacket/pcap"
)

/*
#cgo linux LDFLAGS: -lpcap

#include <sys/socket.h>
#include <sys/ioctl.h>
#include <linux/wireless.h>
#include <pcap.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>

int
is_wireless(const char dev[IFNAMSIZ]) {
	int iswireless = 0, sock_d = -1;
	struct iwreq if_data;

	// Setup the iwreq struct for ioctl call
	memset(&if_data, 0, sizeof(if_data));
	strncpy(if_data.ifr_name, dev, IFNAMSIZ);

	// We require a socket descriptor for the call to ioctl
	// Setup a dummy socket using IPv4 and TCP
	sock_d = socket(AF_INET, SOCK_STREAM, 0);
	if (sock_d == -1) {
		perror("socket");
		return iswireless;
	}

	// wireless protocol == is wireless device
	if (ioctl(sock_d, SIOCGIWNAME, &if_data) != -1) {
			iswireless = 1;
	}

	close(sock_d);
	return (iswireless);
}

int
can_set_rfmon(const char dev[IFNAMSIZ])
{
	int     can_set_rfmon;
	char    errbuf[PCAP_ERRBUF_SIZE];
	pcap_t  *pcap_h;

	can_set_rfmon = 0;

	// create pcap handler for the device.
	pcap_h = pcap_create(dev, errbuf);
	if (pcap_h == NULL) {
			fprintf(stderr, "Fatal error in %s: %s\n",
				"pcap_create", errbuf);
			return (0);
	}

	// check if we can set the device in RF monitor mode.
	if ((can_set_rfmon = pcap_can_set_rfmon(pcap_h)) < 0) {
			fprintf(stderr, "Fatal error in %s: %s\n",
				"pcap_can_set_rfmon", pcap_geterr(pcap_h));
			pcap_close(pcap_h);
			return (0);
	}

	// clean up
	pcap_close(pcap_h);

	return (can_set_rfmon);
}
*/
import "C"

var (
	ErrNoWiDev = errors.New("failed to find any wireless devices")
)

// IsWirelessDevice checks if the given device name is a wireless IO device.
// Returns true if it is, otherwise false.
func IsWirelessDevice(device string) bool {
	return C.is_wireless(C.CString(device)) == 1
}

// CanSetRFMon checks if the given device can be set in RF monitor mode.
// Returns true if it can, otherwise false.
// NOTE: gopacket does not provide a wrapper for pcap_can_set_rfmon, so I did.
func CanSetRFMon(device string) bool {
	return C.can_set_rfmon(C.CString(device)) == 1
}

// FindWirelessDevice attempts to find a wireless IO device and return its
// name.
func FindWirelessDevice() (string, error) {
	// Get list of network devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return "", err
	}
	// For each device, check if it is wireless using our C function
	// NOTE: there isn't really an easy way of checking if a device is wireless
	// in Golang
	for _, device := range devices {
		if IsWirelessDevice(device.Name) {
			return device.Name, nil
		}
	}
	return "", ErrNoWiDev
}
