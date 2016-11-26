package main

import (
  "net"
  "os"
  "path"
  "log"
)

type Agent struct {
	listener net.Listener
	socket   string
}

func NewAgent(socket string) (*Agent, error) {
	a := new(Agent)
  if err := a.Open(socket); err != nil {
    return nil, err
  }
	return a, nil;
}

func (a *Agent) Open(socket string) (error) {
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	a.listener = listener
	a.socket = socket
	if err := os.Chmod(socket, 0600); err != nil {
		a.Close()
		return err
	}
	return nil
}

func (a *Agent) Close() error {
	if err := a.listener.Close(); err != nil {
		return err
	}
	dirname := path.Dir(a.socket)
	if err := os.Remove(dirname); err != nil {
		return err
	}
	return nil
}

func (a *Agent) Run() {
	for {
		fd, err := a.listener.Accept()
		if err != nil {
			break
		}
		go a.Process(fd)
	}
}

func (a *Agent) Process(fd net.Conn) {
	defer fd.Close()
  parser := NewParser()
  if err := parser.Parse(fd); err != nil {
    log.Printf("error: %v", err)
    return
  }
  command := NewCommand()
  res := command.Execute(parser.elements)
  fd.Write(res.data)
}
