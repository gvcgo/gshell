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
	Name  string
	Type  string
	Value string
}

type KtrlContext struct {
	Ctx     *gin.Context
	Command *cobra.Command
	Args    []string
	Options []*Option
}

type KtrlCommand struct {
	Name    string
	Parent  string
	Options []*Option
	RunFunc func(ctx *KtrlContext)
	Handler func(ctx *KtrlContext)
}
