package mcpserver

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
)

var errOutputLimit = errors.New("process output exceeded limit")

type execRunner struct{}

func (execRunner) Run(ctx context.Context, binary string, args []string, dir string, maxOutput int) processResult {
	stdout := newCappedBuffer(maxOutput)
	stderr := newTruncatingBuffer(16 << 10)

	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = dir
	command.Env = []string{
		"PATH=/usr/bin:/bin",
		"LANG=C",
		"LC_ALL=C",
		"TZ=UTC",
	}
	command.Stdout = stdout
	command.Stderr = stderr

	err := command.Run()
	result := processResult{
		stdout:   stdout.Bytes(),
		exitCode: -1,
		err:      err,
	}
	if stdout.exceeded {
		result.outputLimitExceeded = true
		return result
	}
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		result.deadlineExceeded = true
		return result
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		result.exitCode = exitError.ExitCode()
		return result
	}
	if err == nil {
		result.exitCode = 0
	}
	return result
}

type cappedBuffer struct {
	buffer   bytes.Buffer
	maxBytes int
	exceeded bool
}

func newCappedBuffer(maxBytes int) *cappedBuffer {
	return &cappedBuffer{maxBytes: maxBytes}
}

func (buffer *cappedBuffer) Write(data []byte) (int, error) {
	remaining := buffer.maxBytes - buffer.buffer.Len()
	if remaining <= 0 {
		buffer.exceeded = true
		return 0, errOutputLimit
	}
	if len(data) > remaining {
		_, _ = buffer.buffer.Write(data[:remaining])
		buffer.exceeded = true
		return remaining, errOutputLimit
	}
	return buffer.buffer.Write(data)
}

func (buffer *cappedBuffer) Bytes() []byte {
	return bytes.Clone(buffer.buffer.Bytes())
}

type truncatingBuffer struct {
	buffer   bytes.Buffer
	maxBytes int
}

func newTruncatingBuffer(maxBytes int) *truncatingBuffer {
	return &truncatingBuffer{maxBytes: maxBytes}
}

func (buffer *truncatingBuffer) Write(data []byte) (int, error) {
	remaining := buffer.maxBytes - buffer.buffer.Len()
	if remaining > 0 {
		if len(data) > remaining {
			_, _ = buffer.buffer.Write(data[:remaining])
		} else {
			_, _ = buffer.buffer.Write(data)
		}
	}
	return len(data), nil
}
