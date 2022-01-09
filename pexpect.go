package golanglibs

import (
	"os"
	"os/exec"
)

type pexpectStruct struct {
	cmd         *exec.Cmd
	bufall      string // 所有屏幕的显示内容，包括了输入
	ptmx        *os.File
	logToStdout bool // 是否在屏幕打印出整个交互（适合做debug)
	isAlive     bool // 子进程是否在运行
}
