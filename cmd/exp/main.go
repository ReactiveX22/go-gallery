package main

import (
	"errors"
	"fmt"
)

func main() {
	err := CreateUser("test")
	fmt.Println(err)
}

func Connect() error {
	return errors.New("connection Failed")
}

func CreateUser(name string) error {
	err := Connect()
	if err != nil {
		return err
	}
	return nil
}
