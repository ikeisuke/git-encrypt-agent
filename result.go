package main

import(
  "fmt"
  "bytes"
  "strconv"
)

type resultTypeCode int
const (
  NO_TYPE resultTypeCode = iota
	SIMPLE_STRING
  ERROR_STRING
  BINARY_STRING
  MULTIPLE
)

type Result struct {
  resultType resultTypeCode
  data []byte
}

func NewResult() (*Result) {
  r := new(Result)
  return r
}

func (r *Result)SetSimpleString(value string) bool {
  r.resultType = SIMPLE_STRING
  r.data = []byte(fmt.Sprintf("+%v\r\n", value))
  return true
}

func (r *Result)SetErrorString(value string) bool {
  r.resultType = ERROR_STRING
  r.data = []byte(fmt.Sprintf("-%v\r\n", value))
  return true
}

func (r *Result)SetBinaryString(element *Element) bool {
  r.resultType = BINARY_STRING
  buffer := bytes.NewBuffer(make([]byte, 0))
  r.writeBinaryString(buffer, element)
  r.data = buffer.Bytes()
  return true
}

func (r *Result)SetMultiple(elements []*Element) bool {
  r.resultType = MULTIPLE
  length := len(elements)
  buffer := bytes.NewBuffer(make([]byte, 0))
  buffer.Write([]byte("*"))
  buffer.Write([]byte(strconv.Itoa(length)))
  buffer.Write([]byte("\r\n"))
  for i := 0; i < length; i++ {
    element := elements[i]
    r.writeBinaryString(buffer, element)
  }
  r.data = buffer.Bytes()
  return true
}

func (r *Result)writeBinaryString(buffer *bytes.Buffer, element *Element) {
  if element == nil {
    buffer.Write([]byte("$-1\r\n"))
  } else {
    buffer.Write([]byte("$"))
    buffer.Write([]byte(strconv.Itoa(element.size)))
    buffer.Write([]byte("\r\n"))
    buffer.Write(element.data)
    buffer.Write([]byte("\r\n"))
  }
}
