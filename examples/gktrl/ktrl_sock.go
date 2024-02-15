package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gvcgo/gshell/pkgs/ktrl"
	"github.com/gvcgo/gshell/pkgs/shell"
)

func RunKtrlSock() {
	cwd, _ := os.Getwd()
	conf := ktrl.KtrlConf{
		SockDir:  filepath.Join(cwd, "ktrl_dir"),
		SockName: "ktrl_gshell.sock",
	}
	conf.HistoryFilePath = filepath.Join(conf.SockDir, ".history")
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
	time.Sleep(1 * time.Second)
	k.StartShell()
}
