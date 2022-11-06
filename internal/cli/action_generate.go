package cli

import (
	"fmt"

	"github.com/haunt98/populatedb-go/internal/populatedb"
	"github.com/urfave/cli/v2"
)

func (a *action) RunGenerate(c *cli.Context) error {
	a.getFlags(c)

	populator, err := populatedb.NewPopulator(
		a.flags.dialect,
		a.flags.url,
		a.flags.verbose,
		a.flags.dryRun,
	)
	if err != nil {
		return fmt.Errorf("populatedb: failed to new populator: %w", err)
	}

	if err := populator.Insert(c.Context, a.flags.table, a.flags.numberRecord); err != nil {
		return fmt.Errorf("populatedb: failed to insert: %w", err)
	}

	return nil
}
