package shell

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
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
	SetPrompt func(*console.Menu)
	History   readline.History
	cmdList   []*ShellCmd
	flags     map[string][]IShellFlag
}

func NewIShell() (s *IShell) {
	s = &IShell{
		Console: console.New("gshell"),
		flags:   map[string][]IShellFlag{},
		cmdList: []*ShellCmd{},
	}
	s.Console.NewlineBefore = false
	s.Console.NewlineAfter = true
	s.Console.SetPrintLogo(func(c *console.Console) {
		gprint.Yellow("Welcome to gshell!")
	})
	return
}

func (s *IShell) SetupPrompt(setp func(*console.Menu)) {
	s.SetPrompt = setp
}

func (s *IShell) setFlags(command *cobra.Command, opts ...*Flag) {
	if command == nil || len(opts) == 0 {
		return
	}
	command.ResetFlags()
	for _, opt := range opts {
		switch opt.GetType() {
		case OptionTypeBool:
			command.Flags().BoolP(opt.GetName(), opt.GetShort(), gconv.Bool(opt.GetDefault()), opt.GetUsage())
		case OptionTypeInt:
			command.Flags().IntP(opt.GetName(), opt.GetShort(), gconv.Int(opt.GetDefault()), opt.GetUsage())
		case OptionTypeFloat:
			command.Flags().Float64P(opt.GetName(), opt.GetShort(), gconv.Float64(opt.GetDefault()), opt.GetUsage())
		default:
			command.Flags().StringP(opt.GetName(), opt.GetShort(), opt.GetDefault(), opt.GetUsage())
		}
	}
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

		for _, c := range s.cmdList {
			command := &cobra.Command{
				Use:     c.Name,
				Short:   c.HelpStr,
				Long:    c.LongHelpStr,
				GroupID: GroupID,
				Run:     c.Run,
			}
			s.setFlags(command, c.Options...)
			for _, child := range c.Children {
				subCmd := &cobra.Command{
					Use:   child.Name,
					Short: child.HelpStr,
					Long:  child.LongHelpStr,
					Run:   child.Run,
				}
				s.setFlags(subCmd, child.Options...)
				command.AddCommand(subCmd) // add subcommand
			}
			rootCmd.AddCommand(command)
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

func (s *IShell) AddCmd(command *ShellCmd) {
	s.cmdList = append(s.cmdList, command)
}

func (s *IShell) AddChild(parent string, command *ShellCmd) {
	for _, c := range s.cmdList {
		if c.Name == parent {
			c.AddChild(command)
			return
		}
	}
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
