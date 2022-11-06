package cli

import (
	"os"

	"github.com/make-go-great/color-go"
	"github.com/urfave/cli/v2"
)

const (
	name  = "populatedb"
	usage = "fake database data, hehe"

	commandGenerateName  = "generate"
	commandGenerateUsage = "generate fake data then insert to database"

	flagDialectName  = "dialect"
	flagDialectUsage = "database dialect, support mysql for now"

	flagURLName  = "url"
	flagURLUsage = "database url"

	flagTableName  = "table"
	flagTableUsage = "table name to generate data"

	flagNumberRecordName  = "number"
	flagNumberRecordUsage = "number of record to generate"

	flagBatchModeName  = "batch"
	flagBatchModeUsage = "batch mode, insert data in batch"

	flagVerboseName  = "verbose"
	flagVerboseUsage = "show what is going on"

	flagDryRunName  = "dry-run"
	flagDryRunUsage = "demo run without actually changing anything"
)

var (
	commandGenerateAliases = []string{"g", "gen"}
	flagVerboseAliases     = []string{"v"}
)

type App struct {
	cliApp *cli.App
}

func NewApp() *App {
	a := &action{}

	cliApp := &cli.App{
		Name:  name,
		Usage: usage,
		Commands: []*cli.Command{
			{
				Name:    commandGenerateName,
				Aliases: commandGenerateAliases,
				Usage:   commandGenerateUsage,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     flagDialectName,
						Usage:    flagDialectUsage,
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagURLName,
						Usage:    flagURLUsage,
						Required: true,
					},
					&cli.StringFlag{
						Name:     flagTableName,
						Usage:    flagTableUsage,
						Required: true,
					},
					&cli.IntFlag{
						Name:     flagNumberRecordName,
						Usage:    flagNumberRecordUsage,
						Required: true,
					},
					&cli.BoolFlag{
						Name:  flagBatchModeName,
						Usage: flagBatchModeUsage,
					},
					&cli.BoolFlag{
						Name:    flagVerboseName,
						Aliases: flagVerboseAliases,
						Usage:   flagVerboseUsage,
					},
					&cli.BoolFlag{
						Name:  flagDryRunName,
						Usage: flagDryRunUsage,
					},
				},
				Action: a.RunGenerate,
			},
		},
		Action: a.RunHelp,
	}

	return &App{
		cliApp: cliApp,
	}
}

func (a *App) Run() {
	if err := a.cliApp.Run(os.Args); err != nil {
		color.PrintAppError(name, err.Error())
	}
}
