package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/crossedbot/warpath/config"
	database "github.com/crossedbot/warpath/db"
	"github.com/crossedbot/warpath/logger"
	"github.com/crossedbot/warpath/warpath"
)

const (
	FILTER_BEACON_FRAMES     = "type mgt subtype beacon"
	FILTER_PROBE_REQ_FRAMES  = "type mgt subtype probe-req"
	FILTER_PROBE_RESP_FRAMES = "type mgt subtype probe-resp"
)

type Flags struct {
	Device string
	Count  int
}

func main() {
	f := flags()
	c := configuration()
	db, err := database.New(&c)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	warpath_t, err := warpath.New(f.Device, c.SnapLength, (time.Duration(c.Timeout) * time.Second))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := warpath_t.Start(filter(c.Filter)); err != nil {
			log.Fatal(err)
		}
	}()
	loop := f.Count == -1
	for frame := range warpath_t.Output() {
		if err := frame.Save(db); err != nil {
			logger.Log.Error(err)
		}
		if !loop {
			if f.Count <= 1 {
				break
			}
			f.Count--
		}
	}
	warpath_t.Stop()
}

func flags() Flags {
	device := flag.String("device", "", "the wireless device to capture packets")
	count := flag.Int("count", -1, "the number of packets to capture")
	flag.Parse()
	return Flags{
		Device: *device,
		Count:  *count,
	}
}

func configuration() config.Config {
	var c config.Config
	config.Load(&c)
	return c
}

func filter(filterBy string) (filter string) {
	filterBy = strings.ToLower(filterBy)
	switch filterBy {
	case "beacon":
		filter = FILTER_BEACON_FRAMES
	case "probe-req":
		filter = FILTER_PROBE_REQ_FRAMES
	case "probe-resp":
		filter = FILTER_PROBE_RESP_FRAMES
	default:
		filter = fmt.Sprintf("(%s) or (%s) or (%s)",
			FILTER_BEACON_FRAMES,
			FILTER_PROBE_REQ_FRAMES,
			FILTER_PROBE_RESP_FRAMES,
		)
		logger.Info(fmt.Sprintf("using default filter: \"%s\"", filter))
	}
	return
}
