package populatedb

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	tblsschema "github.com/k1LoW/tbls/schema"
)

var ErrNotSupportDatabaseType = errors.New("not support database type")

func ParseDatabaseType(column *tblsschema.Column) (DatabaseType, error) {
	switch {
	case strings.HasPrefix(strings.ToLower(column.Type), "varchar"):
		dtVarchar := DTVarchar{}
		if _, err := fmt.Sscanf(column.Type, "varchar(%d)", &dtVarchar.Length); err != nil {
			return nil, fmt.Errorf("fmt: failed to sscanf [%s]: %w", column.Type, err)
		}

		return &dtVarchar, nil
	case strings.HasPrefix(strings.ToLower(column.Type), "bigint"):
		return &DTBigint{}, nil
	case strings.HasPrefix(strings.ToLower(column.Type), "int"):
		return &DTInt{}, nil
	case strings.EqualFold(column.Type, "timestamp"):
		return &DTTimestamp{}, nil
	case strings.EqualFold(column.Type, "json"):
		return &DTJSON{}, nil
	default:
		return nil, fmt.Errorf("not support database type [%s]: %w", column.Type, ErrNotSupportDatabaseType)
	}
}

type DatabaseType interface {
	Generate() any
}

type DTVarchar struct {
	Length int
}

func (dt *DTVarchar) Generate() any {
	return gofakeit.LetterN(uint(dt.Length))
}

type DTTimestamp struct{}

func (dt *DTTimestamp) Generate() any {
	return time.Now()
}

type DTBigint struct{}

func (dt *DTBigint) Generate() any {
	return gofakeit.Int64()
}

type DTInt struct{}

func (dt *DTInt) Generate() any {
	return gofakeit.Int32()
}

type DTJSON struct{}

func (dt *DTJSON) Generate() any {
	// TODO: need mock
	return nil
}
