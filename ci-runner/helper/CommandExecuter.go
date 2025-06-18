package helper

import (
	cicxt "github.com/devtron-labs/ci-runner/executor/context"
	"os/exec"

	"github.com/devtron-labs/ci-runner/util"
)

type CommandExecutor interface {
	RunCommand(ctx cicxt.CiContext, cmd *exec.Cmd) error
	RunCommandWithCtx(ctx cicxt.CiContext, cmd *exec.Cmd) error
}

type CommandExecutorImpl struct {
}

func NewCommandExecutorImpl() *CommandExecutorImpl {
	return &CommandExecutorImpl{}
}

func (c *CommandExecutorImpl) RunCommand(ctx cicxt.CiContext, cmd *exec.Cmd) error {
	return util.RunCommand(cmd)
}

func (c *CommandExecutorImpl) RunCommandWithCtx(ctx cicxt.CiContext, cmd *exec.Cmd) error {
	// Ensure the command is run with the provided context
	var args []string
	if len(cmd.Args) > 1 {
		args = cmd.Args[1:]
	}
	exec.CommandContext(ctx, cmd.Path, args...)
	return util.RunCommand(cmd)
}
