package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/moqsien/gshell/pkgs/ktrl"
	"github.com/moqsien/gshell/pkgs/shell"
)

func RunKtrlTCP() {
	conf := ktrl.KtrlConf{
		ServerHost: "127.0.0.1",
		ServerPort: 6666,
	}
	cwd, _ := os.Getwd()
	conf.HistoryFilePath = filepath.Join(cwd, "ktrl_dir", ".history")
	conf.MaxHistoryLines = 500

	k := ktrl.NewKtrl(&conf)
	k.AddCommand(&ktrl.KtrlCommand{
		Name:    "show",
		HelpStr: "Show info.",
		Options: []*shell.Flag{
			{
				Name:    "enable",
				Short:   "e",
				Type:    shell.OptionTypeBool,
				Default: "false",
				Usage:   "to enable version[-enable]",
			},
			{
				Name:    "version",
				Short:   "v",
				Type:    shell.OptionTypeString,
				Default: "v0.0.1",
				Usage:   "pass version string[--version=xxx]",
			},
		},
		RunFunc: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			fmt.Println("args info in RunFunc: ", args)
			fmt.Println("Result from server: ", string(ctx.Result))
		},
		Handler: func(ctx *ktrl.KtrlContext) {
			args := ctx.GetArgs()
			fmt.Println("args info in Handler: ", args)
			enable := ctx.GetBool("enable")
			fmt.Println("'enable' in Handler: ", enable)
			version := ctx.GetString("version")
			fmt.Println("'version' in Handler: ", version)

			ctx.SendResponse("hello, ktrl!", 200)
		},
	})
	go k.StartServer()
	time.Sleep(3 * time.Second)
	k.StartShell()
}
