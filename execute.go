package execute

import (
 "time"
 "errors"
 "os/exec"
)

func Timeout(duration int64, name string, arg ...string) error {
	
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
			return errors.New(`Timeout`)
		case err = <- done:
			if err != nil {
				return err
			}
	}
	return nil
	
}
