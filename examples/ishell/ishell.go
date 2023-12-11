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
	ishell.AddCommand(&cobra.Command{
		Use:   "hello",
		Short: "Say hello to you.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("hello, how are you?")
		},
	})
	ishell.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "An example of subcommand.",
	})
	ishell.AddSubCommand("test", &cobra.Command{
		Use:   "show",
		Short: "Show test info.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("show test.")
			fmt.Println(args)
		},
	})
	ishell.SetHistoryFilePath(".gshell_history", 300, true)
	// print logo when shell started.
	ishell.SetPrintLogo(func(_ *console.Console) {
		gprint.Yellow("Welcome to gshell!")
	})
	ishell.Start()
}
