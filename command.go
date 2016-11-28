package main

import (
	"fmt"
	"errors"
)

type Command struct {
	store *Store
}

func NewCommand() *Command {
	c := new(Command)
	c.store = GetSharedStore()
	return c
}

func (c *Command) Execute(elements []*Element) *Result {
	length := len(elements)
	result := NewResult()
	if length == 0 {
		result.SetErrorString("parameter length too short")
		return result
	}
	if length > 3 {
		result.SetErrorString("parameter length too long")
		return result
	}
	name := string(elements[0].data)
	key := ""
	if length >= 2 {
		key = string(elements[1].data)
	}
	value := make([]byte, 0, 0)
	if length == 3 {
		value = elements[2].data
	}
	var err error

	switch name {
	case "set":
		err := c.set(key, value)
		if err == nil {
			result.SetSimpleString("OK")
		} else {
			result.SetErrorString(fmt.Sprintf("%v", err))
		}
	case "get":
		element, err := c.get(key)
		if err == nil {
			result.SetBinaryString(element)
		} else {
			result.SetErrorString(fmt.Sprintf("%v", err))
		}
	case "encrypt":
		element, err := c.encrypt(key, value)
		if err == nil {
			result.SetBinaryString(element)
		} else {
			result.SetErrorString(fmt.Sprintf("%v", err))
		}
	case "decrypt":
		element, err := c.decrypt(key, value)
		if err == nil {
			result.SetBinaryString(element)
		} else {
			result.SetErrorString(fmt.Sprintf("%v", err))
		}
	case "keys":
		elements, err = c.keys()
		if err == nil {
			result.SetMultiple(elements)
		} else {
			result.SetErrorString(fmt.Sprintf("%v", err))
		}
	default:
		result.SetErrorString("command not found")
	}
	return result
}

func (c *Command) set(key string, value []byte) error {
	return c.store.Set(key, value)
}

func (c *Command) get(key string) (*Element, error) {
	value, ok := c.store.Get(key)
	if ok {
		data := []byte(value)
		return NewElement(len(data), data), nil
	}
	return nil, nil
}

func (c *Command) encrypt(key string, data []byte) (*Element, error) {
	value, ok := c.store.Get(key)
	if !ok {
		return nil, errors.New("no encryption keys")
	}
	encryption, err := NewEncryption(value)
	if err != nil {
		return nil, err
	}
	encrypted, err := encryption.encrypt(data)
	if err != nil {
		return nil, err
	}
	return NewElement(len(encrypted), encrypted), nil
}

func (c *Command) decrypt(key string, data []byte) (*Element, error) {
	value, ok := c.store.Get(key)
	if !ok {
		return nil, errors.New("no decryption keys")
	}
	encryption, err := NewEncryption(value)
	if err != nil {
		return nil, err
	}
	encrypted, err := encryption.decrypt(data)
	if err != nil {
		return nil, err
	}
	return NewElement(len(encrypted), encrypted), nil
}

func (c *Command) keys() ([]*Element, error) {
	keys, err := c.store.Keys()
	if err != nil {
		return nil, err
	}
	elements := make([]*Element, len(keys))
	for i := 0; i < len(keys); i++ {
		data := []byte(keys[i])
		element := NewElement(len(data), data)
		elements[i] = element
	}
	return elements, nil
}
