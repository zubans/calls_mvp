package main

import (
	"github.com/zubans/video-call-server/internal/server"
)

func main() {
	// Create and run server
	s := server.NewServer()
	s.Run()
}
