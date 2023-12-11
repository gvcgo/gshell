# gshell是什么
gshell是一个用于创建交互式命令行应用的golang库。它支持与本地、远程的进程进行交互。

# 如何使用

- 简单的交互式应用例子.
  - [ishell](https://github.com/moqsien/gshell/tree/main/examples/ishell)
- 与本地进程进行交互的例子(client和server可以在不同进程运行).
  - [ktrl_sock](https://github.com/moqsien/gshell/blob/main/examples/gktrl/ktrl_sock.go)
- 与远程进程进行交互的例子(client和server可以分别在不同机器上的两个进程中).
  - [ktrl_tcp](https://github.com/moqsien/gshell/blob/main/examples/gktrl/ktrl_tcp.go)

# 许可

MIT
