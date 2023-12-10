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

// Cobra Flags
type Option struct {
	Name    string // flag name
	Type    string // flag type
	Default string // default value
	Usage   string // flag help info
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

// parse flags and args for server.
func (kctx *KtrlContext) parseFlagsArgs() {
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
	kctx.parseFlagsArgs()
	val, _ := kctx.Command.Flags().GetString(name)
	return val
}

func (kctx *KtrlContext) GetBool(name string) bool {
	kctx.parseFlagsArgs()
	val, _ := kctx.Command.Flags().GetBool(name)
	return val
}

func (kctx *KtrlContext) GetInt(name string) int {
	kctx.parseFlagsArgs()
	val, _ := kctx.Command.Flags().GetInt(name)
	return val
}

func (kctx *KtrlContext) GetFloat(name string) float64 {
	kctx.parseFlagsArgs()
	val, _ := kctx.Command.Flags().GetFloat64(name)
	return val
}

type KtrlCommand struct {
	Name        string                 // cmd name
	Parent      string                 // parent cmd name
	HelpStr     string                 // Short for cobra cmd
	LongHelpStr string                 // Long for cobra cmd
	Options     []*Option              // flags for cobra
	RunFunc     func(ctx *KtrlContext) // Not Nil. Hook for cobra.
	Handler     func(ctx *KtrlContext) // Not Nil. Handler for server.
}

// Route for current cmd.
func (kc *KtrlCommand) GetRoute() string {
	if kc.Parent == "" {
		return fmt.Sprintf("/%s/", kc.Name)
	} else {
		return fmt.Sprintf("/%s/%s/", kc.Parent, kc.Name)
	}
}
