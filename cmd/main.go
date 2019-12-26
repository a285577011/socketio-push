package main

import (
	"socketserver/server"
)

func main() {
	s := server.NewServer()
	s.Run()
}
