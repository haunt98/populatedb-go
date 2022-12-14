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

	// 2^16
	maxQuestionMarks = 65536

	stmtInsert = "INSERT INTO %s (%s) VALUES %s;"
)

var (
	ErrNotSupportDialect    = errors.New("not support dialect")
	ErrTableNotExist        = errors.New("table not exist")
	ErrMaximumQuestionMarks = errors.New("maximum question marks")
)

type Populator interface {
	Insert(ctx context.Context, tableName string, numberRecord int) error
	InsertBatch(ctx context.Context, tableName string, numberRecord int) error
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
		return nil, fmt.Errorf("not support dialect [%s]: %w", dbDialect, ErrNotSupportDialect)
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
	columnNames, questionMarks, argFns, err := p.prepareInsert(tableName)
	if err != nil {
		return err
	}

	// INSERT INTO table_name (column1, column2, column3) VALUES (?, ?, ?);
	queryInsert := fmt.Sprintf(stmtInsert,
		tableName,
		strings.Join(columnNames, ", "),
		fmt.Sprintf("(%s)", strings.Join(questionMarks, ", ")),
	)

	for i := 0; i < numberRecord; i++ {
		// Generate each time insert for different value
		args := make([]any, 0, len(argFns))
		for _, argFn := range argFns {
			args = append(args, argFn())
		}

		if p.verbose {
			fmt.Printf("Index: [%d]\n", i)
		}

		if !p.dryRun {
			if _, err := p.db.ExecContext(ctx, queryInsert, args...); err != nil {
				return fmt.Errorf("database: failed to exec [%s]: %w", queryInsert, err)
			}
		}
	}

	return nil
}

func (p *populator) InsertBatch(ctx context.Context, tableName string, numberRecord int) error {
	columnNames, questionMarks, argFns, err := p.prepareInsert(tableName)
	if err != nil {
		return err
	}

	if len(columnNames) == 0 {
		return nil
	}

	numberRecordEachBatch := maxQuestionMarks / len(questionMarks)
	if numberRecordEachBatch == 0 {
		return fmt.Errorf("maximum question marks [%d]: %w", len(questionMarks), ErrMaximumQuestionMarks)
	}

	// Because the numberRecordLastBatch may less than numberRecordEachBatch
	// For example
	// numberRecord := 120
	// numberRecordEachBatch := 50
	// numberBatch := 120/50 + 1 = 3
	// First 2 batches => 50 * 2 = 100
	// Last batch => 120 - 100 = 20
	numberBatch := numberRecord/numberRecordEachBatch + 1
	numberRecordLastBatch := numberRecord - (numberBatch-1)*numberRecordEachBatch

	generateQueryArgsInsertFn := func(tempNumberRecord int) (string, []any) {
		valuesQuestionMarks := make([]string, 0, tempNumberRecord)
		argsInsert := make([]any, 0, tempNumberRecord*len(argFns))
		for i := 0; i < tempNumberRecord; i++ {
			// (?, ?, ?)
			valuesQuestionMarks = append(valuesQuestionMarks, fmt.Sprintf("(%s)", strings.Join(questionMarks, ", ")))

			// Generate each time insert for different value
			args := make([]any, 0, len(argFns))
			for _, argFn := range argFns {
				args = append(args, argFn())
			}
			argsInsert = append(argsInsert, args...)
		}

		// INSERT INTO table_name (column1, column2, column3) VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?);
		queryInsert := fmt.Sprintf(stmtInsert,
			tableName,
			strings.Join(columnNames, ", "),
			strings.Join(valuesQuestionMarks, ", "),
		)

		return queryInsert, argsInsert
	}

	for i := 0; i < numberBatch-1; i++ {
		queryInsert, argsInsert := generateQueryArgsInsertFn(numberRecordEachBatch)

		if p.verbose {
			fmt.Printf("Index: [%d]\n", i)
		}

		if !p.dryRun {
			if _, err := p.db.ExecContext(ctx, queryInsert, argsInsert...); err != nil {
				return fmt.Errorf("database: failed to exec [%s]: %w", queryInsert, err)
			}
		}
	}

	{
		// Last batch
		queryInsert, argsInsert := generateQueryArgsInsertFn(numberRecordLastBatch)

		if p.verbose {
			fmt.Printf("Index: [%d]\n", numberBatch-1)
		}

		if !p.dryRun {
			if _, err := p.db.ExecContext(ctx, queryInsert, argsInsert...); err != nil {
				return fmt.Errorf("database: failed to exec [%s]: %w", queryInsert, err)
			}
		}
	}

	return nil
}

// Return columnNames, questionMarks, argFns
func (p *populator) prepareInsert(tableName string) (columnNames, questionMarks []string, argFns []func() any, err error) {
	table, ok := p.tables[tableName]
	if !ok {
		return nil, nil, nil, fmt.Errorf("table [%s] not exist: %w", tableName, ErrTableNotExist)
	}

	columnNames = make([]string, 0, len(table.Columns))
	questionMarks = make([]string, 0, len(table.Columns))
	argFns = make([]func() any, 0, len(table.Columns))
	for _, column := range table.Columns {
		dt, err := ParseDatabaseType(column)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse database type [%s]: %w", column.Type, err)
		}

		columnNames = append(columnNames, column.Name)
		questionMarks = append(questionMarks, "?")
		argFns = append(argFns, dt.Generate)
	}

	return columnNames, questionMarks, argFns, nil
}
