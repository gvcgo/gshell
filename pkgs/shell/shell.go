package shell

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/reeflective/console"
	"github.com/reeflective/readline"
	"github.com/rsteube/carapace"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	GroupID string = "core"
)

type IShell struct {
	Console   *console.Console
	RootCmd   *cobra.Command
	SetPrompt func(*console.Menu)
	History   readline.History
}

func NewIShell() (s *IShell) {
	s = &IShell{
		Console: console.New("gshell"),
	}
	s.initCommand()
	s.Console.NewlineBefore = false
	s.Console.NewlineAfter = true
	s.Console.SetPrintLogo(func(c *console.Console) {
		gprint.Yellow("Welcome to gshell!")
	})
	return
}

func (s *IShell) initCommand() {
	if s.RootCmd == nil {
		s.RootCmd = &cobra.Command{
			Short: "This is an interactive shell powere by gshell.",
		}
	}
}

func (s *IShell) SetupPrompt(setp func(*console.Menu)) {
	s.SetPrompt = setp
}

func (s *IShell) Start() error {
	// By default the shell as created a single menu and
	// made it current, so you can access it and set it up.
	menu := s.Console.ActiveMenu()

	// Set some custom prompt handlers for this menu.
	if s.SetPrompt == nil {
		s.SetPrompt = SetupPrompt
	}
	s.SetPrompt(menu)

	// history file
	if s.History != nil {
		menu.AddHistorySource("local_history", s.History)
	}

	// We bind a special handler for this menu, which will exit the
	// application (with confirm), when the shell readline receives
	// a Ctrl-D keystroke. You can map any error to any handler.
	menu.AddInterrupt(io.EOF, ExitCtrlD)

	menu.SetCommands(func() *cobra.Command {
		rootCmd := &cobra.Command{
			Short: "This is an interactive shell powered by gshell.",
		}

		rootCmd.AddGroup(&cobra.Group{ID: GroupID, Title: "gshell commands: "})

		// additional commands
		rootCmd.AddCommand(&cobra.Command{
			Use:     "exit",
			Short:   "Exit gshell.",
			GroupID: GroupID,
			Run: func(cmd *cobra.Command, args []string) {
				gprint.Yellow("Exiting...")
				os.Exit(0)
			},
		})

		for _, c := range s.RootCmd.Commands() {
			if c.Name() == "help" {
				continue
			}
			cmd := &cobra.Command{
				Use:     c.Use,
				Short:   c.Short,
				GroupID: GroupID,
				Run:     c.Run,
			}
			for _, sc := range c.Commands() {
				cmd.AddCommand(&cobra.Command{
					Use:   sc.Use,
					Short: sc.Short,
					Run:   sc.Run,
				})
			}
			rootCmd.AddCommand(cmd)
		}

		for _, cmd := range rootCmd.Commands() {
			c := carapace.Gen(cmd)

			if cmd.Args != nil {
				c.PositionalAnyCompletion(
					carapace.ActionCallback(func(c carapace.Context) carapace.Action {
						return carapace.ActionFiles()
					}),
				)
			}

			flagMap := make(carapace.ActionMap)
			cmd.Flags().VisitAll(func(f *pflag.Flag) {
				if f.Name == "file" || strings.Contains(f.Usage, "file") {
					flagMap[f.Name] = carapace.ActionFiles()
				}
			})

			if cmd.Name() == "ssh" {
				// Generate a list of random hosts to use as positional arguments
				hosts := make([]string, 0)
				for i := 0; i < 10; i++ {
					hosts = append(hosts, fmt.Sprintf("host%d", i))
				}
				c.PositionalCompletion(carapace.ActionValues(hosts...))
			}

			if cmd.Name() == "encrypt" {
				cmd.Flags().VisitAll(func(f *pflag.Flag) {
					if f.Name == "algorithm" {
						flagMap[f.Name] = carapace.ActionValues("aes", "des", "blowfish")
					}
				})
			}

			c.FlagCompletion(flagMap)
		}

		rootCmd.SetHelpCommandGroupID(GroupID)
		rootCmd.InitDefaultHelpFlag()
		rootCmd.CompletionOptions.DisableDefaultCmd = true
		rootCmd.DisableFlagsInUseLine = true
		return rootCmd
	})

	err := s.Console.Start()
	return err
}

func (s *IShell) AddCommand(cmds ...*cobra.Command) {
	s.initCommand()
	for _, c := range cmds {
		c.GroupID = GroupID
	}
	s.RootCmd.AddCommand(cmds...)
}

func (s *IShell) AddSubCommand(parent string, cmds ...*cobra.Command) {
	s.initCommand()
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

func (s *IShell) SetHistoryFilePath(fPath string, maxLine int, enableLocal ...bool) {
	// All menus currently each have a distinct, in-memory history source.
	// Replace the main (current) menu's history with one writing to our
	// application history file. The default history is named after its menu.
	if fPath == "" {
		fPath = ".gshell_local_history"
	}
	s.History, _ = EmbeddedHistory(fPath, maxLine, enableLocal...)
}
