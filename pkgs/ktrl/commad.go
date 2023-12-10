package ktrl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

const (
	OptionTypeString  string = "string"
	OptionTypeBool    string = "bool"
	OptionTypeInt     string = "int"
	OptionTypeFloat   string = "float"
	ContextTypeClient int8   = 1
	ContextTypeServer int8   = 2
	QueryArgsName     string = "queryArgs"
)

type Option struct {
	Name    string
	Type    string
	Default string
	Usage   string
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

func (kctx *KtrlContext) parseFlags() {
	if kctx.Type == ContextTypeServer && kctx.Command == nil && kctx.GinCtx != nil {
		kctx.Command = &cobra.Command{}
		for _, opt := range kctx.Options {
			kctx.Command.Flags().Set(opt.Name, kctx.GinCtx.Query(opt.Name))
		}
		argListStr := kctx.GinCtx.Query(QueryArgsName)
		kctx.args = strings.Split(argListStr, ",")
	}
}

func (kctx *KtrlContext) GetArgs() []string {
	return kctx.args
}

func (kctx *KtrlContext) GetString(name string) string {
	kctx.parseFlags()
	val, _ := kctx.Command.Flags().GetString(name)
	return val
}

func (kctx *KtrlContext) GetBool(name string) bool {
	kctx.parseFlags()
	val, _ := kctx.Command.Flags().GetBool(name)
	return val
}

func (kctx *KtrlContext) GetInt(name string) int {
	kctx.parseFlags()
	val, _ := kctx.Command.Flags().GetInt(name)
	return val
}

func (kctx *KtrlContext) GetFloat(name string) float64 {
	kctx.parseFlags()
	val, _ := kctx.Command.Flags().GetFloat64(name)
	return val
}

type KtrlCommand struct {
	Name        string
	Parent      string
	HelpStr     string
	LongHelpStr string
	Options     []*Option
	RunFunc     func(ctx *KtrlContext)
	Handler     func(ctx *KtrlContext)
}

func (kc *KtrlCommand) GetRoute() string {
	if kc.Parent == "" {
		return kc.Name
	} else {
		return fmt.Sprintf("/%s/%s/", kc.Parent, kc.Name)
	}
}
