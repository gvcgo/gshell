package ktrl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/spf13/cobra"
)

type FlagType string

const (
	OptionTypeString  FlagType = "string"
	OptionTypeBool    FlagType = "bool"
	OptionTypeInt     FlagType = "int"
	OptionTypeFloat   FlagType = "float"
	ContextTypeClient int8     = 1
	ContextTypeServer int8     = 2
	QueryArgsName     string   = "queryArgs"
)

// Cobra Flags
type Option struct {
	Name    string   // flag name
	Short   string   // flag shorthand
	Type    FlagType // flag type
	Default string   // default value
	Usage   string   // flag help info
}

type KtrlContext struct {
	GinCtx  *gin.Context
	Command *cobra.Command
	Route   string
	args    []string
	Options []*Option
	Result  []byte
	Type    int8
}

// Send reponse back to client.
func (kctx *KtrlContext) SendResponse(content interface{}, code ...int) {
	if kctx.GinCtx != nil {
		statusCode := http.StatusOK
		if len(code) > 0 {
			statusCode = code[0]
		}
		switch r := content.(type) {
		case string:
			kctx.GinCtx.String(statusCode, r)
		case []byte:
			kctx.GinCtx.String(statusCode, string(r))
		default:
			res, err := json.Marshal(content)
			if err != nil {
				fmt.Println(err)
				kctx.GinCtx.String(http.StatusInternalServerError, err.Error())
				return
			}
			kctx.GinCtx.String(statusCode, string(res))
		}
	}
}

func (kctx *KtrlContext) SetArgs(args ...string) {
	kctx.args = args
}

// parse flags and args for server.
func (kctx *KtrlContext) GetArgs() []string {
	if kctx.Type == ContextTypeServer && kctx.GinCtx != nil {
		args := kctx.GinCtx.Query(QueryArgsName)
		kctx.args = strings.Split(args, ",")
	}
	return kctx.args
}

func (kctx *KtrlContext) GetString(name string) (r string) {
	if kctx.Type == ContextTypeClient {
		r, _ = kctx.Command.Flags().GetString(name)
	} else {
		r = kctx.GinCtx.Query(name)
	}
	return
}

func (kctx *KtrlContext) GetBool(name string) (r bool) {
	if kctx.Type == ContextTypeClient {
		r, _ = kctx.Command.Flags().GetBool(name)
	} else {
		str := kctx.GinCtx.Query(name)
		r = gconv.Bool(str)
	}
	return
}

func (kctx *KtrlContext) GetInt(name string) (r int) {
	if kctx.Type == ContextTypeClient {
		r, _ = kctx.Command.Flags().GetInt(name)
	} else {
		str := kctx.GinCtx.Query(name)
		r = gconv.Int(str)
	}
	return
}

func (kctx *KtrlContext) GetFloat(name string) (r float64) {
	if kctx.Type == ContextTypeClient {
		r, _ = kctx.Command.Flags().GetFloat64(name)
	} else {
		str := kctx.GinCtx.Query(name)
		r = gconv.Float64(str)
	}
	return
}

type KtrlCommand struct {
	Name          string                 // cmd name
	Parent        string                 // parent cmd name
	HelpStr       string                 // Short for cobra cmd
	LongHelpStr   string                 // Long for cobra cmd
	SendInRunFunc bool                   // Send request in RunFunc
	Options       []*Option              // flags for cobra
	RunFunc       func(ctx *KtrlContext) // Not Nil. Hook for cobra.
	Handler       func(ctx *KtrlContext) // Not Nil. Handler for server.
}

// Route for current cmd.
func (kc *KtrlCommand) GetRoute() string {
	return FormatRoute(kc.Name, kc.Parent)
}

func FormatRoute(name, parent string) string {
	if parent == "" {
		return fmt.Sprintf("/%s/", name)
	} else {
		return fmt.Sprintf("/%s/%s/", parent, name)
	}
}
