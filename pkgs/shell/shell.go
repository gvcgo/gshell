package shell

import (
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

type IShell struct {
	Console *console.Console
	RootCmd *cobra.Command
}
