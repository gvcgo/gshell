package ktrl

const (
	SockName string = "ktrl_ishell.sock"
)

type KtrlConf struct {
	SockPath        string
	ServerPort      int
	ServerIP        string
	HistoryFilePath string
}
