package ktrl

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/gshell/pkgs/shell"
	"github.com/reeflective/console"
	"github.com/spf13/cobra"
)

const (
	PingRoute string = "/ping"
)

type Ktrl struct {
	iShell   *shell.IShell
	client   *http.Client
	engine   *gin.Engine
	conf     *KtrlConf
	commands []*KtrlCommand
	l        *sync.Mutex
}

func NewKtrl(cfg *KtrlConf) (k *Ktrl) {
	k = &Ktrl{
		conf: cfg,
		l:    &sync.Mutex{},
	}
	return k
}

func (k *Ktrl) AddCommand(kcmd *KtrlCommand) {
	k.l.Lock()
	k.commands = append(k.commands, kcmd)
	k.l.Unlock()
}

/*
client
*/
func (k *Ktrl) parseParams(params map[string]string) (p string) {
	for k, v := range params {
		if len(p) == 0 {
			p += fmt.Sprintf("?%s=%s", k, v)
		} else {
			p += fmt.Sprintf("&%s=%s", k, v)
		}
	}
	return
}

func (k *Ktrl) getResult(ctx *KtrlContext) {
	if k.client == nil {
		if k.conf.SockDir != "" && k.conf.SockName != "" {
			// Unix socket
			k.client = &http.Client{}
			k.client.Transport = &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", filepath.Join(k.conf.SockDir, k.conf.SockName))
				},
			}
		} else if k.conf.ServerPort != 0 && k.conf.ServerHost != "" {
			// TCP
			k.client = &http.Client{}
		} else {
			return
		}
	}
	params := map[string]string{}
	for _, opt := range ctx.Options {
		switch opt.Type {
		case OptionTypeBool:
			v, _ := ctx.Command.Flags().GetBool(opt.Name)
			params[opt.Name] = gconv.String(v)
		case OptionTypeInt:
			v, _ := ctx.Command.Flags().GetInt(opt.Name)
			params[opt.Name] = gconv.String(v)
		case OptionTypeFloat:
			v, _ := ctx.Command.Flags().GetFloat64(opt.Name)
			params[opt.Name] = gconv.String(v)
		default:
			v, _ := ctx.Command.Flags().GetString(opt.Name)
			params[opt.Name] = v
		}
	}
	if len(ctx.args) > 0 {
		params[QueryArgsName] = strings.Join(ctx.args, ",")
	}

	var kUrl string
	if k.conf.SockDir != "" {
		kUrl = fmt.Sprintf("http://%s%s%s", k.conf.SockName, ctx.Route, k.parseParams(params))
	} else {
		kUrl = fmt.Sprintf("http://%s:%d%s%s", k.conf.ServerHost, k.conf.ServerPort, ctx.Route, k.parseParams(params))
	}
	resp, err := k.client.Get(kUrl)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	ctx.Result, _ = io.ReadAll(resp.Body)
}

func (k *Ktrl) addShellCmd() {
	for _, c := range k.commands {
		if c.RunFunc == nil {
			continue
		}
		icmd := &cobra.Command{
			Use:   c.Name,
			Short: c.HelpStr,
			Long:  c.LongHelpStr,
			Run: func(cmd *cobra.Command, args []string) {
				ctx := &KtrlContext{
					Command: cmd,
					args:    args,
					Options: c.Options,
					Route:   c.GetRoute(),
					Type:    ContextTypeClient,
				}
				k.getResult(ctx)
				c.RunFunc(ctx)
			},
		}

		for _, opt := range c.Options {
			switch opt.Type {
			case OptionTypeBool:
				if opt.Short == "" {
					icmd.Flags().Bool(opt.Name, gconv.Bool(opt.Default), opt.Usage)
				} else {
					icmd.Flags().BoolP(opt.Name, opt.Short, gconv.Bool(opt.Default), opt.Usage)
				}
			case OptionTypeInt:
				if opt.Short == "" {
					icmd.Flags().Int(opt.Name, gconv.Int(opt.Default), opt.Usage)
				} else {
					icmd.Flags().IntP(opt.Name, opt.Short, gconv.Int(opt.Default), opt.Usage)
				}
			case OptionTypeFloat:
				if opt.Short == "" {
					icmd.Flags().Float64(opt.Name, gconv.Float64(opt.Default), opt.Usage)
				} else {
					icmd.Flags().Float64P(opt.Name, opt.Short, gconv.Float64(opt.Default), opt.Usage)
				}
			default:
				if opt.Short == "" {
					icmd.Flags().String(opt.Name, opt.Default, opt.Usage)
				} else {
					icmd.Flags().StringP(opt.Name, opt.Short, opt.Default, opt.Usage)
				}
			}
		}
		if c.Parent == "" {
			k.iShell.AddCommand(icmd)
		} else {
			k.iShell.AddSubCommand(c.Parent, icmd)
		}
	}
}

func (k *Ktrl) SetPrintLogo(f func(_ *console.Console)) {
	k.iShell.SetPrintLogo(f)
}

func (k *Ktrl) StartShell() error {
	if k.iShell == nil {
		k.iShell = shell.NewIShell()
		k.iShell.SetHistoryFilePath(k.conf.HistoryFilePath, k.conf.MaxHistoryLines, true)
	}
	k.addShellCmd()
	err := k.iShell.Start()
	return err
}

/*
server
*/
func (k *Ktrl) initEngine() {
	if k.engine == nil {
		gin.SetMode(gin.ReleaseMode)
		k.engine = gin.New()
	}
}

func (k *Ktrl) addServerHandlers() {
	k.initEngine()
	for _, c := range k.commands {
		ctx := &KtrlContext{
			Route:   c.GetRoute(),
			Options: c.Options,
			Type:    ContextTypeServer,
		}
		k.engine.GET(ctx.Route, func(gctx *gin.Context) {
			ctx.GinCtx = gctx
			c.Handler(ctx)
		})
	}
	// Check if server is running.
	k.engine.GET(PingRoute, func(gctx *gin.Context) {
		gctx.JSON(200, gin.H{"message": "pong"})
	})
}

func (k *Ktrl) checkAndRemoveOldSockFile(sockPath string) {
	if ok, _ := gutils.PathIsExist(sockPath); ok {
		os.RemoveAll(sockPath)
	}
}

func (k *Ktrl) listen() error {
	k.initEngine()
	var listener net.Listener
	if k.conf.SockDir != "" && k.conf.SockName != "" {
		sockPath := filepath.Join(k.conf.SockDir, k.conf.SockName)
		k.checkAndRemoveOldSockFile(sockPath)
		unixAddr, err := net.ResolveUnixAddr("unix", sockPath)
		if err != nil {
			return err
		}
		listener, err = net.ListenUnix("unix", unixAddr)
		if err != nil {
			return err
		}
	} else if k.conf.ServerHost != "" && k.conf.ServerPort != 0 {
		var err error
		// listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", k.conf.ServerHost, k.conf.ServerPort))
		listener, err = net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", k.conf.ServerPort))
		if err != nil {
			return err
		}
	}
	return http.Serve(listener, k.engine)
}

func (k *Ktrl) StartServer() {
	k.initEngine()
	k.addServerHandlers()
	k.listen()
}
