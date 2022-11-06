package populatedb

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-sql-driver/mysql"
	tblsconfig "github.com/k1LoW/tbls/config"
	tblsdatasource "github.com/k1LoW/tbls/datasource"
	tblsschema "github.com/k1LoW/tbls/schema"
)

const (
	dialectMySQL = "mysql"
)

var ErrDialectNotSupport = errors.New("dialect not support ")

type Populator interface{}

type populator struct {
	db         *sql.DB
	tblsSchema *tblsschema.Schema
}

func NewPopulator(dbDialect, dbURL string) (Populator, error) {
	if dbDialect != dialectMySQL {
		return nil, fmt.Errorf("not support [%s]: %w", dbDialect, ErrDialectNotSupport)
	}

	// https://go.dev/doc/tutorial/database-access
	mysqlCfg, err := mysql.ParseDSN(dbURL)
	if err != nil {
		return nil, fmt.Errorf("mysql: failed to parse dsn [%s]: %w", dbURL, err)
	}

	// https://github.com/go-sql-driver/mysql#timetime-support
	mysqlCfg.ParseTime = true
	mysqlCfg.AllowNativePasswords = true
	mysqlCfg.Loc = time.UTC

	mysqlURL := mysqlCfg.FormatDSN()
	db, err := sql.Open(dbDialect, mysqlURL)
	if err != nil {
		return nil, fmt.Errorf("sql: failed to open [%s]: %w", mysqlURL, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database: failed to ping [%s] : %w", mysqlURL, err)
	}

	// https://github.com/k1LoW/tbls
	// https://stackoverflow.com/q/48671938
	tblsURL := "mysql://" + mysqlCfg.User + ":" + url.QueryEscape(mysqlCfg.Passwd) + "@" + mysqlCfg.Addr + "/" + mysqlCfg.DBName
	tblsSchema, err := tblsdatasource.Analyze(tblsconfig.DSN{
		URL: tblsURL,
	})
	if err != nil {
		return nil, fmt.Errorf("tbls: faield to analyze [%s]: %w", tblsURL, err)
	}

	return &populator{
		db:         db,
		tblsSchema: tblsSchema,
	}, nil
}
