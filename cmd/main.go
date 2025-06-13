package main

import (
	"github.com/Parapheen/ph-clone/internal/server"
)

const (
	addr = ":3333"
)

func main() {
	s := server.NewServer()

	s.Run(addr)
}

