package main

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/teeworlds-go/econ"
)

func asyncWriteLine(ctx context.Context, wg *sync.WaitGroup, conn *econ.Conn, commandChan <-chan string) {
	defer func() {
		wg.Done()
		log.Println("command writer closed")
	}()

	var err error
	for {
		select {
		case <-ctx.Done():
			log.Printf("closing command writer: %v", ctx.Err())
			return
		case command, ok := <-commandChan:
			if !ok {
				log.Println("command channel closed")
				return
			}
			err = conn.WriteLine(command)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Printf("closing command writer: %v", ctx.Err())
					return
				}
				log.Printf("failed to write line: %v", err)
			}
		}
	}
}
