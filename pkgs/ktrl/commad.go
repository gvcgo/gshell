package ktrl

import (
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

const (
	OptionTypeString string = "string"
	OptionTypeBool   string = "bool"
	OptionTypeInt    string = "int"
	OptionTypeFloat  string = "float"
)

type Option struct {
	Name    string
	Type    string
	Default string
	Usage   string
}

type KtrlContext struct {
	Ctx     *gin.Context
	Command *cobra.Command
	Route   string
	Args    []string
	Options []*Option
	Result  []byte
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
