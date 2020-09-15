package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	beaconFilter := filter("beacon")
	require.Equal(t, FILTER_BEACON_FRAMES, beaconFilter)
	probeReqFilter := filter("probe-req")
	require.Equal(t, FILTER_PROBE_REQ_FRAMES, probeReqFilter)
	probeRespFilter := filter("probe-resp")
	require.Equal(t, FILTER_PROBE_RESP_FRAMES, probeRespFilter)
}
