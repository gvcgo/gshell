package main

import (
	"fmt"

	"github.com/moqsien/gshell/pkgs/shell"
	"github.com/spf13/cobra"
)

func main() {
	ishell := shell.NewIShell()
	ishell.AddCommand(&cobra.Command{
		Use: "hello",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("hello, how are you?")
		},
	})
	ishell.AddCommand(&cobra.Command{
		Use: "test",
	})
	ishell.AddSubCommand("test", &cobra.Command{
		Use: "show",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("show test.")
		},
	})

	ishell.Start()
}
