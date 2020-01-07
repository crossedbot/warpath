package db

import (
	"database/sql"

	"bitbucket.org/liamstask/goose/lib/goose"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"

	"github.com/crossedbot/warpath/config"
	"github.com/crossedbot/warpath/logger"
)

func New(conf *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(conf.Database.Name, conf.Database.Path)
	if err != nil {
		return nil, err
	}
	logger.SetFile(conf.Logging.File)
	db.LogMode(conf.Logging.Mode)
	db.SetLogger(logger.Log)
	db.DB().SetMaxOpenConns(conf.Database.MaxOpenConnections)
	if err := migrate(
		db.DB(),
		conf.Database.Name,
		conf.Database.Path,
		conf.Database.MigrationsPath,
		conf.Database.MigrationsEnv,
	); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func driver(name, openStr string) goose.DBDriver {
	d := goose.DBDriver{
		Name:    name,
		OpenStr: openStr,
	}
	switch name {
	case "sqlite3":
		d.Import = "github.com/mattn/go-sqlite3"
		d.Dialect = &goose.Sqlite3Dialect{}
	}
	return d
}

func migrate(db *sql.DB, dbname, dbpath, migrationsdir, migrationsenv string) error {
	c := &goose.DBConf{
		MigrationsDir: migrationsdir,
		Env:           migrationsenv,
		Driver:        driver(dbname, dbpath),
	}
	v, err := goose.GetMostRecentDBVersion(migrationsdir)
	if err != nil {
		return err
	}
	return goose.RunMigrationsOnDb(c, migrationsdir, v, db)
}
