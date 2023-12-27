package shell

import "github.com/spf13/cobra"

type ShellCmd struct {
	Name        string  // cmd name
	Parent      string  // parent cmd name
	HelpStr     string  // Short for cobra cmd
	LongHelpStr string  // Long for cobra cmd
	Options     []*Flag // flags for cobra
	Run         func(cmd *cobra.Command, args []string)
	Children    []*ShellCmd
}

func NewShellCmd() (sc *ShellCmd) {
	sc = &ShellCmd{
		Children: []*ShellCmd{},
	}
	return
}

func (s *ShellCmd) AddChild(child *ShellCmd) {
	s.Children = append(s.Children, child)
}
