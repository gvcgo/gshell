package ktrl

const (
	SockName string = "ktrl_ishell.sock"
)

type KtrlConf struct {
	SockDir         string // unix socket file directory
	SockName        string // unix socket file name
	ServerPort      int    // remote server port
	ServerHost      string // remote server host
	HistoryFilePath string // gshell history file path
	MaxHistoryLines int    // max history lines to store
}
