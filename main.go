package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	database "github.com/crossedbot/common/golang/db"
	"github.com/crossedbot/common/golang/logger"
	"github.com/crossedbot/common/golang/service"
	"github.com/crossedbot/warpath/warpath"
)

const (
	// Capturing filters
	FILTER_BEACON_FRAMES     = "type mgt subtype beacon"
	FILTER_PROBE_REQ_FRAMES  = "type mgt subtype probe-req"
	FILTER_PROBE_RESP_FRAMES = "type mgt subtype probe-resp"

	// Exit codes
	FATAL_EXITCODE = iota + 1
)

// Flags represent the program flags.
type Flags struct {
	Device string // Local Wifi device
	Count  int    // Number of packets to capture
}

func main() {
	ctx := context.Background()
	svc := service.New(ctx)
	if err := svc.Run(run, syscall.SIGINT, syscall.SIGTERM); err != nil {
		fatal("Error: %s", err)
	}
}

// fatal log the formatted message as an error and exits on FATAL_EXITCODE.
func fatal(format string, a ...interface{}) {
	logger.Error(fmt.Errorf(format, a...))
	os.Exit(1)
}

// flags returns the program's flags.
func flags() Flags {
	device := flag.String("device", "", "the wireless device to capture packets")
	count := flag.Int("count", -1, "the number of packets to capture")
	flag.Parse()
	return Flags{
		Device: *device,
		Count:  *count,
	}
}

// configuration returns the loaded program configuration.
func configuration() Config {
	var c Config
	Load(&c)
	return c
}

// newDB returns a new database based on the given configuration.
func newDB(c *Config) (database.Database, error) {
	db, err := database.New(
		c.Database.Name,
		c.Database.Path,
		c.Database.MaxOpenConnections,
	)
	if err != nil {
		return nil, err
	}
	db.LogMode(c.Logging.Mode)
	db.SetLogger(logger.Log)
	if err := db.Migrate(
		c.Database.Name,
		c.Database.Path,
		c.Database.MigrationsPath,
		c.Database.MigrationsEnv,
	); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// filter returns the filter used for capturing packets; current options are:
//	- beacon
//	- probe-req
//	- probe-resp
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

// run is the entrypoint into the warpath program.
func run(ctx context.Context) error {
	// setup
	f := flags()
	cfg := configuration()
	logger.SetFile(cfg.Logging.File)
	db, err := newDB(&cfg)
	if err != nil {
		return err
	}
	defer db.Close()
	// initialize a new warpath handle
	warpath_t, err := warpath.New(
		f.Device,
		cfg.SnapLength,
		(time.Duration(cfg.Timeout) * time.Second),
	)
	if err != nil {
		return err
	}
	defer warpath_t.Close()
	// start the packet capturing
	out, err := warpath_t.Run(filter(cfg.Filter))
	if err != nil {
		return err
	}
	// by default loop forever, otherwise loop until count number of packets are
	// captured
	loop := f.Count == -1
	for frame := range out {
		if err := db.SaveTx(frame); err != nil {
			logger.Error(err)
		}
		if !loop {
			if f.Count <= 1 {
				break
			}
			f.Count--
		}
	}
	return nil
}
