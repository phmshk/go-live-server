package utils

import (
	"errors"
	"fmt"
	"log"
	"net"
	"syscall"
)

func GetAvailablePort(startPort int) (net.Listener, error) {
	port := startPort
	for {
		address := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			return listener, nil
		}
		if !isPortInUse(err) {
			return nil, err
		}

		log.Printf("Port busy, retrying on %d\n", port+1)
		port++
	}
}

func isPortInUse(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, syscall.EADDRINUSE)
}
