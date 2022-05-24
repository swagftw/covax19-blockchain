package main

import (
	"log"

	"github.com/swagftw/covax19-blockchain/transport"
)

func main() {
	errChan := make(chan error)
	// get echo
	ech := transport.InitEcho()

	// init handlers
	transport.InitHandlers(ech)

	// start server
	go func() {
		err := transport.StartHTTPServer(ech)
		if err != nil {
			errChan <- err
		}
	}()

	// Wait for the server to exit
	err := <-errChan
	log.Println(err)
}
