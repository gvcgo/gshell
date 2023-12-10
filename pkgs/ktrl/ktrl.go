package ktrl

import (
	"github.com/gin-gonic/gin"
	"github.com/moqsien/gshell/pkgs/shell"
)

type Ktrl struct {
	iShell   *shell.IShell
	router   *gin.Engine
	conf     *KtrlConf
	commands []*KtrlCommand
}
