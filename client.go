package main

import (
  "net"
  "errors"
  "log"
  "fmt"
)

type Client struct {
  elements []*Element
  conn net.Conn
}

func NewClient(socket string) (*Client, error) {
  c := new(Client);
  if err := c.Open(socket); err != nil {
    return nil, err
  }
  return c, nil
}

func (c *Client) Open(socket string) error {
  conn, err := net.Dial("unix", socket)
  if err != nil {
    log.Printf("error: %v\n", err)
    return err
  }
  c.conn = conn
  return nil
}

func (c *Client) Set(name string, key []byte) error {
  c.addData([]byte("set"))
  c.addData([]byte(name))
  c.addData(key)
  return nil
}

func (c *Client) Encrypt(name string, plaintext []byte) error {
  c.addData([]byte("encrypt"))
  c.addData([]byte(name))
  c.addData(plaintext)
  return nil
}

func (c *Client) Decrypt(name string, ciphertext []byte) error {
  c.addData([]byte("decrypt"))
  c.addData([]byte(name))
  c.addData(ciphertext)
  return nil
}

func (c *Client) Send() ([]byte, error) {
  req := NewResult();
  if req.SetMultiple(c.elements) {
    _, err := c.conn.Write(req.data)
    if err != nil {
      return nil, err
    }
    err = c.conn.(*net.UnixConn).CloseWrite()
    if err != nil {
      return nil, err
    }
    parser := NewParser()
    parser.Parse(c.conn)
    element := parser.elements[0]
    if element.kind == ERROR_STRING {
      return nil, errors.New(fmt.Sprintf("ERROR: %v", string(parser.elements[0].data)))
    } else {
      return parser.elements[0].data, nil
    }
  }
  return nil, errors.New("unknown error")
}

func (c *Client) addData(data []byte) error{
  c.elements = append(c.elements, NewElement(len(data), data));
  return nil
}
