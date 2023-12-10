package shell

import (
	"io"
	"os"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
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
		gprint.Yellow("Welcome to gshell!")
	})
	return
}

func (s *IShell) InitCommand() {
	if s.RootCmd == nil {
		s.RootCmd = &cobra.Command{}
	}
	s.AddCommand(&cobra.Command{
		Use:   "exit",
		Short: "Exit gshell.",
		Run: func(cmd *cobra.Command, args []string) {
			gprint.Yellow("Exiting...")
			os.Exit(0)
		},
	})
}

func (s *IShell) Start() error {
	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := s.Console.ActiveMenu()

	// Set some custom prompt handlers for this menu.
	SetupPrompt(menu)

	// We bind a special handler for this menu, which will exit the
	// application (with confirm), when the shell readline receives
	// a Ctrl-D keystroke. You can map any error to any handler.
	menu.AddInterrupt(io.EOF, ExitCtrlD)

	menu.SetCommands(func() *cobra.Command {
		s.RootCmd.InitDefaultHelpCmd()
		s.RootCmd.CompletionOptions.DisableDefaultCmd = true
		s.RootCmd.DisableFlagsInUseLine = true
		return s.RootCmd
	})

	err := s.Console.Start()
	return err
}

func (s *IShell) AddCommand(cmds ...*cobra.Command) {
	if s.RootCmd == nil {
		return
	}
	s.RootCmd.AddCommand(cmds...)
}

func (s *IShell) AddSubCommand(parent string, cmds ...*cobra.Command) {
	if s.RootCmd == nil {
		return
	}
	for _, cmd := range s.RootCmd.Commands() {
		if cmd.Name() == parent {
			cmd.AddCommand(cmds...)
			return
		}
	}

	s.RootCmd.AddCommand(&cobra.Command{
		Use: parent,
	})
	s.AddSubCommand(parent, cmds...)
}

func (s *IShell) SetPrintLogo(f func(_ *console.Console)) {
	s.Console.SetPrintLogo(f)
}

func (s *IShell) SetHistoryFilePath(fPath string) {
	menu := s.Console.ActiveMenu()
	// All menus currently each have a distinct, in-memory history source.
	// Replace the main (current) menu's history with one writing to our
	// application history file. The default history is named after its menu.
	if fPath == "" {
		fPath = ".gshell_local_history"
	}
	hist, _ := EmbeddedHistory(fPath)
	menu.AddHistorySource("local_history", hist)
}
