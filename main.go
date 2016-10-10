package main

import (
  "os"
  "os/signal"
  "io/ioutil"
  "fmt"
  "net"
  "syscall"
  "strings"
  "strconv"
  "path/filepath"
  "log"
)

var encryption_keys map[string]string = map[string]string{}

func server(listener net.Listener) {
  for {
    fd, err := listener.Accept()
    if err != nil {
      //log.Printf("error: %v\n", err)
      return
    }
    ch := make(chan string)
    go func(ch chan string) {
      input := ""
      for {
        buf := make([]byte, 512)
        nr, err := fd.Read(buf)
        if err != nil {
          break
        }
        data := buf[0:nr]
        input += string(data)
      }
      splited := strings.SplitN(strings.Trim(input, "\n"), " ", 3)
      size := len(splited)
      command := splited[0]
      output := ""
      switch command {
      case "set":
        if (size == 3) {
          key := splited[1]
          value := splited[2]
          output = setEncryptionKey(key, value)
        } else {
          output = "ERR: Not enough input"
        }
      case "get":
        if (size >= 2) {
          key := splited[1]
          output = getEncrptionKey(key)
        } else {
          output = "ERR: Not enough input"
        }
      case "list":
        output = listEncryptionKeys()
      default:
        output = "ERR: command not found"
      }
      _, err = fd.Write([]byte(output))
      if err != nil {
        log.Printf("error: %v\n", err)
      }
      fd.Close()
      ch <- "finish"
    }(ch)
    _, _ = <- ch
  }
}

func setEncryptionKey(key string, value string) string{
  encryption_keys[key] = value
  return "OK"
}

func getEncrptionKey(key string) string {
  v, ok := encryption_keys[key]
  if (ok) {
    return v
  }
  return "(nil)"
}

func listEncryptionKeys() string {
  list := []string{}
  for key,value := range encryption_keys {
    list = append(list, key + ":" + value)
  }
  return strings.Join(list, "\n")
}

func main() {
  log.SetFlags(log.Lshortfile)
  encrypt_socket_dir, err := ioutil.TempDir("", "git-encrypt-agent.")
  pid := strconv.Itoa(os.Getpid())
  encrypt_socket_file := encrypt_socket_dir + "/agent." + pid
  listener, err := net.Listen("unix", encrypt_socket_file)
  if err != nil {
    log.Printf("error: %v\n", err)
    return
  }
  if err := os.Chmod(encrypt_socket_file, 0700); err != nil {
    fmt.Println(err)
  }
  clean := make(chan int)
  shutdown(listener, clean)
  fmt.Println(fmt.Sprintf("GIT_ENCRYPT_SOCK=%v;export GIT_ENCRYPT_SOCK;", encrypt_socket_file))
  fmt.Println(fmt.Sprintf("GIT_ENCRYPT_PID=%v;export GIT_ENCRYPT_PID;", pid))
  server(listener)
  _ = <-clean
}

func shutdown(listener net.Listener, clean chan int) {
  c := make(chan os.Signal, 2)
  signal.Notify(c, os.Interrupt, syscall.SIGTERM)
  go func() {
    interrupt := 0
    for {
      s := <-c
      switch s {
      case os.Interrupt, syscall.SIGINT:
        if (interrupt == 0) {
          fmt.Println("Interrupt...")
          interrupt++
          continue
        }
      }
      break
    }
    file, err := listener.(*net.UnixListener).File()
    if (err != nil) {
      log.Printf("error: %v\n", err)
    }
    if err := listener.Close(); err != nil {
      log.Printf("error: %v\n", err)
    }
    dirname := strings.Replace(filepath.Dir(file.Name()), "unix:", "", 1)
    if err := os.RemoveAll(dirname); err != nil {
      log.Printf("error: %v\n", err)
    }
    clean <- 1
  }()
}
