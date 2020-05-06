package steamcmd

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

type SteamCmd struct {
	cmd               *exec.Cmd
	Available, Active bool
	pipe              *bufio.ReadWriter
}

func NewSteamCmd(root string) (*SteamCmd, error) {
	cmd := exec.Command("bash", root, "+login", "anonymous")
	cmd.Stderr = os.Stderr
	pr, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	wc, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	pipe := bufio.NewReadWriter(bufio.NewReader(pr), bufio.NewWriter(wc))
	return &SteamCmd{cmd: cmd, pipe: pipe, Available: false, Active: false}, nil
}

func (s *SteamCmd) Run() error {
	go s.cmd.Run()
	for {
		line, err := s.pipe.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				continue
			}
			return err
		}
		if strings.Contains(line, "Waiting for user info...OK") {
			s.Available = true
			return nil
		}
	}
}

func (s *SteamCmd) RunScript(script string) (result []byte, err error) {
	if !s.Available && !s.Active {
		err = errors.New("service not available")
	}
	log.Println("run script : " + script)
	if _, err = s.pipe.WriteString(script + "\n"); err != nil {
		return
	}
	if err = s.pipe.Flush(); err != nil {
		return
	}
	s.Active = true

	defer func() {
		s.Active = false
	}()

	var token = 0
	for {
		line, _, err := s.pipe.ReadLine()
		if err != nil {
			if err == io.EOF {
				continue
			}
			return nil, err
		}
		if token == 1 {
			result = line
			break
		}
		if strings.Contains(string(line), "Steam>") {
			token++
		}
	}
	return
}

func (s *SteamCmd) Close() error {
	if s.Available && !s.Active {
		_, err := s.RunScript("quit")
		return err
	} else {
		if s.cmd.Process != nil {
			return s.cmd.Process.Kill()
		}
	}
	return nil
}
