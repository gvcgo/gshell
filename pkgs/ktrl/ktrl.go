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
	PingRoute    string = "/ping/"
	PingResponse string = "pong"
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

func (k *Ktrl) GetShell() (sh *shell.IShell) {
	return k.iShell
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

func (k *Ktrl) getClient() {
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
}

func (k *Ktrl) GetResult(ctx *KtrlContext) {
	k.getClient()
	if k.client == nil {
		return
	}
	params := map[string]string{}
	if ctx.Command != nil {
		for _, opt := range ctx.Options {
			switch opt.GetType() {
			case shell.OptionTypeBool:
				v, _ := ctx.Command.Flags().GetBool(opt.GetName())
				params[opt.GetName()] = gconv.String(v)
			case shell.OptionTypeInt:
				v, _ := ctx.Command.Flags().GetInt(opt.GetName())
				params[opt.GetName()] = gconv.String(v)
			case shell.OptionTypeFloat:
				v, _ := ctx.Command.Flags().GetFloat64(opt.GetName())
				params[opt.GetName()] = gconv.String(v)
			default:
				v, _ := ctx.Command.Flags().GetString(opt.GetName())
				params[opt.GetName()] = v
			}
		}
	} else {
		for _, opt := range ctx.Options {
			params[opt.GetName()] = opt.GetDefault()
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
		command := c // replicate, in case "c" will be covered.
		if command.RunFunc == nil {
			continue
		}

		shellCmd := shell.NewShellCmd()
		shellCmd.Name = command.Name
		shellCmd.HelpStr = command.HelpStr
		shellCmd.LongHelpStr = command.LongHelpStr
		shellCmd.Options = command.Options
		shellCmd.Run = func(cmd *cobra.Command, args []string) {
			ctx := &KtrlContext{
				Command: cmd,
				args:    args,
				Options: c.Options,
				Route:   command.GetRoute(),
				Type:    ContextTypeClient,
			}
			if !command.SendInRunFunc {
				k.GetResult(ctx)
			}
			command.RunFunc(ctx)
		}
		if command.Parent == "" {
			k.iShell.AddCmd(shellCmd)
		} else {
			k.iShell.AddChild(c.Parent, shellCmd)
		}
	}
}

// Send msg to server manually.
func (k *Ktrl) SendMsg(name, parent string, options []*shell.Flag, args ...string) (r []byte) {
	ctx := &KtrlContext{
		Command: nil,
		args:    args,
		Options: options,
		Route:   FormatRoute(name, parent),
		Type:    ContextTypeClient,
	}
	k.GetResult(ctx)
	return ctx.Result
}

func (k *Ktrl) SetPrintLogo(f func(_ *console.Console)) {
	k.iShell.SetPrintLogo(f)
}

func (k *Ktrl) PreShellStart() {
	if k.iShell == nil {
		k.iShell = shell.NewIShell()
		k.iShell.SetHistoryFilePath(k.conf.HistoryFilePath, k.conf.MaxHistoryLines, true)
	}
	k.addShellCmd()
}

func (k *Ktrl) StartShell() error {
	if k.iShell == nil {
		k.PreShellStart()
	}
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
		command := c // replicate, in case "c" will be covered.
		ctx := &KtrlContext{
			Route:   command.GetRoute(),
			Options: command.Options,
			Type:    ContextTypeServer,
		}
		k.engine.GET(ctx.Route, func(gctx *gin.Context) {
			ctx.GinCtx = gctx
			command.Handler(ctx)
		})
	}
	// Check if server is running.
	k.engine.GET(PingRoute, func(gctx *gin.Context) {
		gctx.JSON(200, gin.H{"message": PingResponse})
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

func (k *Ktrl) PreServerStart() {
	if k.engine == nil {
		gin.SetMode(gin.ReleaseMode)
		k.engine = gin.New()
	}
	k.addServerHandlers()
}

func (k *Ktrl) StartServer() {
	if k.engine == nil {
		k.PreServerStart()
	}
	k.listen()
}
