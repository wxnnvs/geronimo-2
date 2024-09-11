# Geronimo 2

A Remote Administration Tool for Windows written in Go

## How to Build

To build this project, follow the steps below:

1. Clone the repository to your local machine.
2. Run `go build server.go`
3. Run `server.exe` and port forward port 8080
4. Specify your ip and port in `client.go`
5. Run `go build -ldflags="-H windowsgui" client.go`
6. Run ``client.exe``

That's it! You should now have remote command prompt access 

## Commands:
```
- list               --> list all connected clients by id
- run <id> <command> --> run a command on the specified target
```