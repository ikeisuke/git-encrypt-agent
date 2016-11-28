package main

import (
	"github.com/codegangsta/cli"
	"os"
	"os/signal"
	"os/exec"
	"io/ioutil"
	"strconv"
	"syscall"
	"fmt"
	"log"
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
	}
	app.Run(os.Args)
}

func startAgent(c *cli.Context) (error) {
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

func runAgent(socket string) error {
	agent, err := NewAgent(socket)
	if (err != nil) {
		return err;
	}
	notify := make(chan int)
	signalHandler(notify);
	go func() {
		_ = <-notify;
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
