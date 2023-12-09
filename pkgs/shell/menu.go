package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/reeflective/console"
)

// setupPrompt is a function which sets up the prompts for the main menu.
func SetupPrompt(m *console.Menu) {
	p := m.Prompt()

	p.Primary = func() string {
		prompt := "\x1b[33mexample\x1b[0m [main] in \x1b[34m%s\x1b[0m\n> "
		wd, _ := os.Getwd()

		dir, err := filepath.Rel(os.Getenv("HOME"), wd)
		if err != nil {
			dir = filepath.Base(wd)
		}

		return fmt.Sprintf(prompt, dir)
	}

	p.Secondary = func() string { return ">" }
	p.Right = func() string {
		return "\x1b[1;30m" + time.Now().Format("03:04:05") + "\x1b[0m"
	}

	p.Transient = func() string { return "\x1b[1;30m" + ">> " + "\x1b[0m" }
}
