package execute

import (
 "time"
 "errors"
 "os/exec"
 "github.com/AlasdairF/Custom"
 "io"
)

var ErrTimeout = errors.New(`Timeout`)

type cmd struct {
	duration time.Duration
	name string
	arg []string
}

func Timeout(duration time.Duration, name string, arg ...string) error {
	
	cmd := exec.Command(name, arg...)
	err := cmd.Start()
	if err != nil {
		return err
	}
	
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
		case <- time.After(duration):
			if err = cmd.Process.Kill(); err != nil {
				return err
			}
			<- done // allow goroutine to exit
			return ErrTimeout
		case err = <- done:
			if err != nil {
				return err
			}
	}
	return nil
	
}

func Command(duration time.Duration, name string, arg ...string) *cmd {
	return &cmd{duration, name, arg}
}

func (c *cmd) Run() error {
	return Timeout(c.duration, c.name, c.arg)
}

func (c *cmd) CombinedOutput() ([]byte, error) {
	
	cmd := exec.Command(c.name, c.arg...)
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer stderr.Close()
	
	b := custom.NewBuffer(0)
	defer b.Close()
	
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	
	go io.Copy(b, stdout)
    go io.Copy(b, stderr)
	
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
		case <- time.After(c.duration):
			if err = cmd.Process.Kill(); err != nil {
				return nil, err
			}
			<- done // allow goroutine to exit
			return nil, ErrTimeout
		case err = <- done:
			if err != nil {
				return nil, err
			}
	}
	
	return b.BytesCopy(), nil
	
}
