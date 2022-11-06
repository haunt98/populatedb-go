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

	flagNumberRecordName  = "number"
	flagNumberRecordUsage = "number of record to generate"

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
	cliApp := &cli.App{}

	return &App{
		cliApp: cliApp,
	}
}

func (a *App) Run() {
	if err := a.cliApp.Run(os.Args); err != nil {
		color.PrintAppError(name, err.Error())
	}
}
