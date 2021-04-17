package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

type Cmd struct {
	// cmd           *exec.Cmd
	command       []string
	cwd           string
	env           []string
	stdin         io.Reader
	stdoutWriters []io.Writer
	stderrWriters []io.Writer
	wait          bool
	timeout       time.Duration
	preStartFunc  func() error
	postStartFunc func() error
	postEndFunc   func() error
}

func NewCommand(name string) *Cmd {
	return &Cmd{
		command: []string{name},
		wait:    true,
		env:     os.Environ(),
	}
}

func (c Cmd) Base() *Cmd {
	c.command = []string{c.command[0]}
	return &c
}

func (c Cmd) Command() []string {
	return c.command
}

func (c *Cmd) Args(args ...string) *Cmd {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		c.command = append(c.command, arg)
	}
	return c
}

func (c *Cmd) Arg(name, v string) *Cmd {
	if name == "" || v == "" {
		return c
	}
	c.command = append(c.command, name, v)
	return c
}

func (c *Cmd) ListArg(name string, vs []string) *Cmd {
	if name == "" {
		return c
	}
	for _, v := range vs {
		if v == "" {
			continue
		}
		c.command = append(c.command, name, v)
	}
	return c
}

func (c *Cmd) Cwd(path string) *Cmd {
	c.cwd = path
	return c
}

func (c *Cmd) Env(env ...string) *Cmd {
	c.env = append(c.env, env...)
	return c
}

func (c *Cmd) Stdin(r io.Reader) *Cmd {
	c.stdin = r
	return c
}

func (c *Cmd) Stdout(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stdoutWriters = append(c.stdoutWriters, w)
	return c
}

func (c *Cmd) Stderr(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stderrWriters = append(c.stderrWriters, w)
	return c
}

func (c *Cmd) Timeout(t time.Duration) *Cmd {
	c.timeout = t
	return c
}

func (c *Cmd) NoWait() *Cmd {
	c.wait = false
	return c
}

func (c *Cmd) PreStart(f func() error) *Cmd {
	c.preStartFunc = f
	return c
}

func (c *Cmd) PostStart(f func() error) *Cmd {
	c.postStartFunc = f
	return c
}

func (c *Cmd) PostEnd(f func() error) *Cmd {
	c.postEndFunc = f
	return c
}

func (c *Cmd) Run() (string, error) {
	cmd := exec.Command(c.command[0], c.command[1:]...)
	cmd.Env = append(os.Environ(), c.env...)
	if c.cwd != "" {
		cmd.Dir = c.cwd
	}
	logrus.Tracef("executing %q", cmd.String())

	combinedBuf := new(bytes.Buffer)
	traceWriter := logrus.New().WriterLevel(logrus.TraceLevel)
	warnWriter := logrus.New().WriterLevel(logrus.WarnLevel)

	c.stdoutWriters = append(c.stdoutWriters, combinedBuf, traceWriter)
	c.stderrWriters = append(c.stderrWriters, combinedBuf, warnWriter)
	cmd.Stdout = io.MultiWriter(c.stdoutWriters...)
	cmd.Stderr = io.MultiWriter(c.stderrWriters...)

	if c.preStartFunc != nil {
		if err := c.preStartFunc(); err != nil {
			return "", err
		}
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	if c.postStartFunc != nil {
		if err := c.postStartFunc(); err != nil {
			return "", err
		}
	}

	if c.wait {
		if c.timeout != 0 {
			timer := time.AfterFunc(c.timeout, func() {
				_ = cmd.Process.Kill()
			})
			defer timer.Stop()
		}

		if err := cmd.Wait(); err != nil {
			if c.timeout == 0 {
				return "", err
			}
		}
		if c.postEndFunc != nil {
			if err := c.postEndFunc(); err != nil {
				return "", err
			}
		}
	}

	return combinedBuf.String(), nil
}
