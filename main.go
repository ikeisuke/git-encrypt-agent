package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/codegangsta/cli"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

const PROJECT_NAME = "git-encrypt"
const SOCKET_FILE_NAME = "agent"

func main() {
	log.SetFlags(log.Lshortfile)
	app := cli.NewApp()
	app.Name = "git-encrypt-agent"
	app.Usage = "git enctyption key management agent"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:    "start",
			Aliases: []string{"s"},
			Usage:   "start git-encrypt-agent daemon",
			Action:  startAgent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "a",
					Value: "$TMPDIR/git-encrypt-XXXXXXXXXX/agent.<ppid>",
					Usage: "Bind the agent to the UNIX-domain socket bind_address.",
				},
				cli.IntFlag{
					Name:  "t",
					Value: 0,
					Usage: "Set a default value for the maximum lifetime of identities added to the agent.",
				},
				cli.BoolFlag{
					Name:  "d",
					Usage: "Debug mode. When this option is specified git-encrypt-agent will not daemonize and will write debug information to standard error.",
				},
			},
		},
		{
			Name:   "daemon",
			Action: daemonizeAgent,
			Hidden: true,
		},
		{
			Name:    "stop",
			Aliases: []string{"k"},
			Usage:   "stop git-encrypt-agent daemon",
			Action:  stopAgent,
		},
		{
			Name:    "add",
			Aliases: []string{"a"},
			Usage:   "add key for encrypt and decrypt",
			Action:  addKey,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "Name for encryption key.",
				},
			},
		},
		{
			Name:    "get",
			Aliases: []string{"g"},
			Usage:   "get hash for encrypt and decrypt key",
			Action:  getHash,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "Name for encryption key.",
				},
			},
		},
		{
			Name:    "encrypt",
			Aliases: []string{"e"},
			Usage:   "encrypt data",
			Action:  encrypt,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "Name for encryption key.",
				},
			},
		},
		{
			Name:    "decrypt",
			Aliases: []string{"d"},
			Usage:   "encrypt data",
			Action:  decrypt,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name",
					Usage: "Name for encryption key.",
				},
			},
		},
	}
	app.Run(os.Args)
}

func startAgent(c *cli.Context) error {
	if c.Bool("d") {
		return runAgent()
	} else {
		cmd := exec.Command(os.Args[0], "daemon")
		if err := cmd.Start(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
	}
	return nil
}

func stopAgent(c *cli.Context) error {
	pidFile, err := pidFile(false)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	pidtmp, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	pid, err := strconv.Atoi(string(pidtmp))
	process, err := os.FindProcess(pid)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func daemonizeAgent(c *cli.Context) error {
	if err := runAgent(); err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func runAgent() error {
	socket, err := socketFile(true)
	if err != nil {
		return err
	}
	agent, err := NewAgent(socket)
	if err != nil {
		return err
	}
	notify := make(chan int)
	signalHandler(notify)
	go func() {
		_ = <-notify
		if err := agent.Close(); err != nil {
			log.Printf("error: %v", err)
		}
	}()
	_, err = pidFile(true)
	if err != nil {
		return err
	}
	defer func() {
		dir, _ := tmpDir(false)
		os.RemoveAll(dir)
	}()
	agent.Run()
	return nil
}

func addKey(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return cli.NewExitError("argument --name is required", 1)
	}
	stdin := os.Stdin
	buffer := bytes.NewBuffer([]byte(""))
	for {
		buf := make([]byte, 512)
		nr, err := stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		buffer.Write(buf[0:nr])
	}
	stdin.Close()
	key := buffer.Bytes()
	if len(key) != 32 {
		return cli.NewExitError(fmt.Sprintf("Error: invalid key size %v", len(key)), 1)
	}
	socket, err := socketFile(false)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.Set(name, key)
	res, err := client.Send()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(res)
	writer.Write([]byte{10})
	writer.Flush()
	return nil
}

func getHash(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return cli.NewExitError("argument --name is required", 1)
	}
	socket, err := socketFile(false)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.GetHash(name)
	res, err := client.Send()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(res)
	writer.Write([]byte{10})
	writer.Flush()
	return nil
}

func encrypt(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return cli.NewExitError("argument --name is required", 1)
	}
	stdin := os.Stdin
	buffer := bytes.NewBuffer([]byte(""))
	for {
		buf := make([]byte, 512)
		nr, err := stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		buffer.Write(buf[0:nr])
	}
	stdin.Close()
	socket, err := socketFile(false)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.Encrypt(name, buffer.Bytes())
	encrypted, err := client.Send()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(encrypted)
	writer.Flush()
	return nil
}

func decrypt(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return cli.NewExitError("argument --name is required", 1)
	}
	stdin := os.Stdin
	buffer := bytes.NewBuffer([]byte(""))
	for {
		buf := make([]byte, 512)
		nr, err := stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		buffer.Write(buf[0:nr])
	}
	stdin.Close()
	if len(buffer.Bytes()) == 0 {
		return nil
	}
	socket, err := socketFile(false)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.Decrypt(name, buffer.Bytes())
	encrypted, err := client.Send()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(encrypted)
	writer.Flush()
	return nil
}

func socketFile(creates bool) (string, error) {
	tempDir, err := tmpDir(creates)
	if err != nil {
		return "", err
	}
	socket := tempDir + "/" + PROJECT_NAME + ".sock"
	return socket, nil
}

func pidFile(creates bool) (string, error) {
	tempDir, err := tmpDir(creates)
	if err != nil {
		return "", err
	}
	pidFile := tempDir + "/" + PROJECT_NAME + ".pid"
	if creates {
		pid := strconv.Itoa(os.Getpid())
		err := ioutil.WriteFile(pidFile, []byte(pid), 0600)
		if err != nil {
			return "", err
		}
	}
	return pidFile, nil
}

func tmpDir(creates bool) (string, error) {
	dir := os.Getenv("TMPDIR")
	if len(dir) == 0 {
		dir = "/tmp"
	}
	dir = strings.TrimSuffix(dir, "/")
	executor, err := user.Current()
	if err != nil {
		return "", err
	}
	tmpdir := fmt.Sprintf("%v/%v.%v", dir, PROJECT_NAME, executor.Uid)
	if creates {
		_, err := os.Stat(tmpdir)
		if err != nil {
			if err := os.Mkdir(tmpdir, 0700); err != nil {
				return "", err
			}
		}
	}
	return tmpdir, nil
}

func signalHandler(notify chan int) {
	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		interrupt := 0
		for {
			s := <-sigs
			switch s {
			case os.Interrupt, syscall.SIGINT:
				if interrupt == 0 {
					fmt.Println("Interrupt...")
					interrupt++
					continue
				}
			}
			notify <- 1
			break
		}
	}()
}
