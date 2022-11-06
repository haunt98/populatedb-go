package populatedb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	tblsconfig "github.com/k1LoW/tbls/config"
	tblsdatasource "github.com/k1LoW/tbls/datasource"
	tblsschema "github.com/k1LoW/tbls/schema"
)

const (
	dialectMySQL = "mysql"

	stmtInsert = "INSERT INTO %s (%s) VALUES (%s);"
)

var (
	ErrNotSupportDialect = errors.New("not support dialect")
	ErrTableNotExist     = errors.New("table not exist")
)

type Populator interface {
	Insert(ctx context.Context, tableName string, numberRecord int) error
}

type populator struct {
	db         *sql.DB
	tblsSchema *tblsschema.Schema
	tables     map[string]*tblsschema.Table
	verbose    bool
	dryRun     bool
}

func NewPopulator(
	dbDialect string,
	dbURL string,
	verbose bool,
	dryRun bool,
) (Populator, error) {
	if dbDialect != dialectMySQL {
		return nil, fmt.Errorf("not support [%s]: %w", dbDialect, ErrNotSupportDialect)
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

	tables := make(map[string]*tblsschema.Table, len(tblsSchema.Tables))
	for _, table := range tblsSchema.Tables {
		tables[table.Name] = table
	}

	return &populator{
		db:         db,
		tblsSchema: tblsSchema,
		tables:     tables,
		verbose:    verbose,
		dryRun:     dryRun,
	}, nil
}

func (p *populator) Insert(ctx context.Context, tableName string, numberRecord int) error {
	table, ok := p.tables[tableName]
	if !ok {
		return fmt.Errorf("table [%s] not exist: %w", tableName, ErrTableNotExist)
	}

	columnNames := make([]string, 0, len(table.Columns))
	questionMarks := make([]string, 0, len(table.Columns))
	argFns := make([]func() any, 0, len(table.Columns))
	for _, column := range table.Columns {
		dt, err := ParseDatabaseType(column.Type)
		if err != nil {
			return fmt.Errorf("failed to parse database type [%s]: %w", column.Type, err)
		}

		columnNames = append(columnNames, column.Name)
		questionMarks = append(questionMarks, "?")
		argFns = append(argFns, dt.Generate)
	}

	queryInsert := fmt.Sprintf(stmtInsert,
		tableName,
		strings.Join(columnNames, ", "),
		strings.Join(questionMarks, ", "),
	)

	for i := 0; i < numberRecord; i++ {
		args := make([]any, 0, len(argFns))
		for _, argFn := range argFns {
			args = append(args, argFn())
		}

		if p.verbose {
			fmt.Println(i, queryInsert, args)
		}

		if !p.dryRun {
			if _, err := p.db.ExecContext(ctx, queryInsert, args...); err != nil {
				return fmt.Errorf("database: failed to exec [%s]: %w", queryInsert, err)
			}
		}
	}

	return nil
}
