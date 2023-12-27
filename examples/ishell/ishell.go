package main

import (
	"fmt"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/gshell/pkgs/shell"
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

func main() {
	ishell := shell.NewIShell()

	h := shell.NewShellCmd()
	h.Name = "hello"
	h.HelpStr = "Say hello to you."
	h.Options = []*shell.Flag{
		{
			Name:    "enable",
			Short:   "e",
			Usage:   "enable extra.",
			Default: "false",
			Type:    shell.OptionTypeBool,
		},
	}
	h.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("hello, how are you?")
		enable, _ := cmd.Flags().GetBool("enable")
		if enable {
			fmt.Println("Extra info is enabled.")
		}
	}
	ishell.AddCmd(h)

	tParent := "test"
	t := shell.NewShellCmd()
	t.Name = tParent
	t.HelpStr = "An example of subcommand."
	ishell.AddCmd(t)

	sub := shell.NewShellCmd()
	sub.Name = "show"
	sub.HelpStr = "Show test info."
	sub.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("show test.")
		fmt.Println(args)
	}
	ishell.AddChild(tParent, sub)

	ishell.SetHistoryFilePath(".gshell_history", 300, true)
	// print logo when shell started.
	ishell.SetPrintLogo(func(_ *console.Console) {
		gprint.Yellow("Welcome to gshell!")
	})
	ishell.Start()
}
