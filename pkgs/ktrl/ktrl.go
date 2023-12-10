package ktrl

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/moqsien/gshell/pkgs/shell"
	"github.com/spf13/cobra"
)

const (
	ArgsPattern string = "args_%s"
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
			k.client = &http.Client{}
			k.client.Transport = &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", filepath.Join(k.conf.SockDir, k.conf.SockName))
				},
			}
		} else if k.conf.ServerPort != 0 && k.conf.ServerHost != "" {
			k.client = &http.Client{}
		} else {
			return
		}
	}
	params := map[string]string{}
	for _, opt := range ctx.Options {
		switch opt.Type {
		case "bool":
			v, _ := ctx.Command.Flags().GetBool(opt.Name)
			params[opt.Name] = gconv.String(v)
		case "int":
			v, _ := ctx.Command.Flags().GetInt(opt.Name)
			params[opt.Name] = gconv.String(v)
		case "float":
			v, _ := ctx.Command.Flags().GetFloat64(opt.Name)
			params[opt.Name] = gconv.String(v)
		default:
			v, _ := ctx.Command.Flags().GetString(opt.Name)
			params[opt.Name] = v
		}
	}
	if len(ctx.Args) > 0 {
		params[fmt.Sprintf(ArgsPattern, ctx.Command.Name())] = strings.Join(ctx.Args, ",")
	}

	var kUrl string
	if k.conf.SockDir != "" {
		kUrl = fmt.Sprintf("http://%s/%s/%s", k.conf.SockName, ctx.Route, k.parseParams(params))
	} else {
		kUrl = fmt.Sprintf("http://%s:%d/%s/%s", k.conf.ServerHost, k.conf.ServerPort, ctx.Route, k.parseParams(params))
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
				route := c.Name
				if c.Parent != "" {
					route = c.Parent + "/" + c.Name
				}
				ctx := &KtrlContext{
					Command: cmd,
					Args:    args,
					Options: c.Options,
					Route:   route,
				}
				k.getResult(ctx)
				c.RunFunc(ctx)
			},
		}

		for _, opt := range c.Options {
			switch opt.Type {
			case "bool":
				icmd.Flags().Bool(opt.Name, false, opt.Usage)
			case "int":
				icmd.Flags().Int(opt.Name, 0, opt.Usage)
			case "float":
				icmd.Flags().Float64(opt.Name, 0, opt.Usage)
			default:
				icmd.Flags().String(opt.Name, opt.Default, opt.Usage)
			}
		}
		if c.Parent == "" {
			k.iShell.AddCommand(icmd)
		} else {
			k.iShell.AddSubCommand(c.Parent, icmd)
		}
	}
}

func (k *Ktrl) StartShell() {
	if k.iShell == nil {
		k.iShell = shell.NewIShell()
	}
	k.addShellCmd()
	k.iShell.Start()
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
	// TODO: add handlers
}

func (k *Ktrl) listen() {
	k.initEngine()
	// TODO: listen
}

func (k *Ktrl) StartServer() {
	k.initEngine()
	k.addServerHandlers()
	k.listen()
}
