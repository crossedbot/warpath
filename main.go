package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinzhu/gorm"

	database "github.com/crossedbot/common/golang/db"
	"github.com/crossedbot/common/golang/logger"
	"github.com/crossedbot/warpath/config"
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
	logger.SetFile(c.Logging.File)
	db, err := newDB(&c)
	if err != nil {
		fatal("failed to create database: %s", err.Error())
	}
	defer db.Close()
	warpath_t, err := warpath.New(f.Device, c.SnapLength, (time.Duration(c.Timeout) * time.Second))
	if err != nil {
		fatal("failed to create Warpath handler: %s", err.Error())
	}
	go func() {
		if err := warpath_t.Start(filter(c.Filter)); err != nil {
			fatal("failed to start capture: %s", err.Error())
		}
	}()
	loop := f.Count == -1
	for frame := range warpath_t.Output() {
		if err := frame.Save(db); err != nil {
			logger.Error(err)
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

func fatal(format string, a ...interface{}) {
	logger.Error(fmt.Errorf(format, a...))
	os.Exit(1)
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

func newDB(c *config.Config) (*gorm.DB, error) {
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
	if err := database.Migrate(
		db.DB(),
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
