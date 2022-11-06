package cli

import (
	"log"

	"github.com/urfave/cli/v2"
)

type action struct {
	flags struct {
		dialect      string
		url          string
		table        string
		numberRecord int
		verbose      bool
		dryRun       bool
	}
}

func (a *action) RunHelp(c *cli.Context) error {
	return cli.ShowAppHelp(c)
}

func (a *action) getFlags(c *cli.Context) {
	a.flags.dialect = c.String(flagDialectName)
	a.flags.url = c.String(flagURLName)
	a.flags.table = c.String(flagTableName)
	a.flags.numberRecord = c.Int(flagNumberRecordName)

	a.flags.verbose = c.Bool(flagVerboseName)
	a.flags.dryRun = c.Bool(flagDryRunName)

	a.log("Flags %+v\n", a.flags)
}

func (a *action) log(format string, v ...interface{}) {
	if a.flags.verbose {
		log.Printf(format, v...)
	}
}
