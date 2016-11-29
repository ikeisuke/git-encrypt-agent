package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"bytes"
	"io"
	"bufio"
)

const PROJECT_NAME = "git-encrypt"
const SOCKET_FILE_NAME = "agent"

func main() {
	log.SetFlags(log.Lshortfile)
	app := cli.NewApp()
	app.Name = "git-encrypt-agent"
	app.Usage = "git enctyption key management agent"
	app.Version = "0.0.1"
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
				cli.StringFlag{
					Name:  "key",
					Usage: "Data of encryption key.",
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
	socket, err := socketFile()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	pid := os.Getpid()
	if !c.Bool("d") {
		cmd := exec.Command(os.Args[0], "daemon")
		cmd.Env = append(os.Environ(), fmt.Sprintf("GIT_ENCRYPT_SOCK=%v", socket))
		if err := cmd.Start(); err != nil {
			return cli.NewExitError(err.Error(), 1)
		}
		pid = cmd.Process.Pid
	}
	fmt.Printf("GIT_ENCRYPT_SOCK=%v;export GIT_ENCRYPT_SOCK;\n", socket)
	fmt.Printf("GIT_ENCRYPT_PID=%v;export GIT_ENCRYPT_PID;\n", pid)
	fmt.Printf("echo Agent pid %v;\n", pid)
	if c.Bool("d") {
		return runAgent(socket)
	}
	return nil
}

func stopAgent(c *cli.Context) error {
	pid, err := strconv.Atoi(os.Getenv("GIT_ENCRYPT_PID"))
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	err = process.Kill()
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	return nil
}

func daemonizeAgent(c *cli.Context) error {
	os.Stdin.Close()
	os.Stdout.Close()
	os.Stderr.Close()
	socket := os.Getenv("GIT_ENCRYPT_SOCK")
	if socket == "" {
		return cli.NewExitError("GIT_ENCRYPT_SOCK not set, cannot run agent", 1)
	}
	return runAgent(socket)
}

func addKey(c *cli.Context) error {
	name := c.String("name")
	if name == "" {
		return cli.NewExitError("argument --name is required", 1)
	}
	key := c.String("key")
	if key == "" {
		return cli.NewExitError("argument --key is required", 1)
	}
	if len(key) != 32 {
		return cli.NewExitError(fmt.Sprintf("argument --key is require 32 charactors, now %v", len(key)), 1)
	}
	socket := os.Getenv("GIT_ENCRYPT_SOCK")
	if socket == "" {
		return cli.NewExitError("GIT_ENCRYPT_SOCK not set, cannot run agent", 1)
	}
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.Set(name, []byte(key))
	res, err := client.Send();
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
	socket := os.Getenv("GIT_ENCRYPT_SOCK")
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.Encrypt(name, buffer.Bytes())
	encrypted, err := client.Send();
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
	socket := os.Getenv("GIT_ENCRYPT_SOCK")
	client, err := NewClient(socket)
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	client.Decrypt(name, buffer.Bytes())
	encrypted, err := client.Send();
	if err != nil {
		return cli.NewExitError(err.Error(), 1)
	}
	writer := bufio.NewWriter(os.Stdout)
	writer.Write(encrypted)
	writer.Flush()
	return nil
}

func runAgent(socket string) error {
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
	agent.Run()
	return nil
}

func socketFile() (string, error) {
	tempDir, err := ioutil.TempDir("", fmt.Sprintf("%s.", PROJECT_NAME))
	if err != nil {
		return "", err
	}
	pid := strconv.Itoa(os.Getpid())
	socket := tempDir + "/" + SOCKET_FILE_NAME + "." + pid
	if err := os.Chmod(tempDir, 0700); err != nil {
		return "", err
	}
	return socket, nil
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
