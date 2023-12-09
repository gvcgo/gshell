package shell

import (
	"io"

	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

type IShell struct {
	Console *console.Console
	RootCmd *cobra.Command
}

func NewIShell() (s *IShell) {
	s = &IShell{
		Console: console.New("gshell"),
	}
	s.InitCommand()
	s.Console.NewlineBefore = true
	s.Console.NewlineAfter = true
	s.Console.SetPrintLogo(func(c *console.Console) {

	})
	return
}

func (s *IShell) InitCommand() {
	if s.RootCmd == nil {
		s.RootCmd = &cobra.Command{}
	}
	// TODO: some commands by default
}

func (s *IShell) Start() error {
	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := s.Console.ActiveMenu()

	// Set some custom prompt handlers for this menu.
	SetupPrompt(menu)

	// All menus currently each have a distinct, in-memory history source.
	// Replace the main (current) menu's history with one writing to our
	// application history file. The default history is named after its menu.
	hist, _ := EmbeddedHistory(".example-history")
	menu.AddHistorySource("local history", hist)

	// We bind a special handler for this menu, which will exit the
	// application (with confirm), when the shell readline receives
	// a Ctrl-D keystroke. You can map any error to any handler.
	menu.AddInterrupt(io.EOF, ExitCtrlD)

	menu.SetCommands(func() *cobra.Command {
		return s.RootCmd
	})

	err := s.Console.Start()
	return err
}

func (s *IShell) AddCommand(cmds ...*cobra.Command) {
	if s.RootCmd != nil {
		s.RootCmd.AddCommand(cmds...)
	}
}

func (s *IShell) AddSubCommand(parent string, cmds ...*cobra.Command) {
	if s.RootCmd == nil {
		return
	}
	for _, cmd := range s.RootCmd.Commands() {
		if cmd.Name() == parent {
			cmd.AddCommand(cmds...)
		}
	}
}
