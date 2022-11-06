package populatedb

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

var ErrNotSupportDatabaseType = errors.New("not support database type")

// varchar(123)
// timestamp
func ParseDatabaseType(databaseTypeStr string) (DatabaseType, error) {
	switch {
	case strings.HasPrefix(strings.ToLower(databaseTypeStr), "varchar"):
		dtVarchar := DTVarchar{}
		if _, err := fmt.Sscanf(databaseTypeStr, "varchar(%d)", &dtVarchar.Length); err != nil {
			return nil, fmt.Errorf("fmt: failed to sscanf [%s]: %w", databaseTypeStr, err)
		}

		return &dtVarchar, nil
	case strings.HasPrefix(strings.ToLower(databaseTypeStr), "bigint"):
		return &DTBigint{}, nil
	case strings.HasPrefix(strings.ToLower(databaseTypeStr), "int"):
		return &DTInt{}, nil
	case strings.EqualFold(databaseTypeStr, "timestamp"):
		return &DTTimestamp{}, nil
	case strings.EqualFold(databaseTypeStr, "json"):
		return &DTJSON{}, nil
	default:
		return nil, fmt.Errorf("not support database type [%s]: %w", databaseTypeStr, ErrNotSupportDatabaseType)
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
