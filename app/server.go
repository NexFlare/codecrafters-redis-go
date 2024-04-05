package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

var (
	storedData store.Store = store.GetStore()
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	port := flag.String("port", "6379", "port to listen on")
	flag.Parse()
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", *port))
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		defer c.Close()
		go func() {
			buffer := make([]byte, 128)
			for {
				_, err = c.Read(buffer)
				if err != nil {
					fmt.Println("Error reading command: ", err.Error())
					return
				}
				fmt.Println("CMD is", string(buffer))
				req := string(buffer)
				handleCommand(c, req)
				buffer =  make([]byte, 128)
			}
			
		}()
	}
	
	
}

func handleCommand(c net.Conn, data string) {
	
	cmd := command.NewCommand(data)

	if cmd == nil {
		c.Write([]byte(getSimpleString("ERROR")))
		return
	}

	switch cmd.Command {
	case command.PING:
		c.Write([]byte(getSimpleString("PONG")))
	case command.ECHO:
		str := strings.Join(cmd.Arguments, " ")
		c.Write([]byte(getSimpleString(str)))
	case command.SET:
		if len(cmd.Arguments) == 2 {
			storedData.Set(cmd.Arguments[0], cmd.Arguments[1])
			c.Write([]byte(getSimpleString("OK")))
		} else if len(cmd.Arguments) == 4 {
			duration, err := strconv.Atoi(cmd.Arguments[3])
			if err != nil {
				c.Write([]byte(getSimpleString("ERROR")))
				return
			}
			storedData.SetWithExpiry(cmd.Arguments[0], cmd.Arguments[1], int64(duration))
			c.Write([]byte(getSimpleString("OK")))
		} else {
			c.Write([]byte(getSimpleString("")))
		}
	case command.GET:
		str := storedData.Get(cmd.Arguments[0])
		c.Write([]byte(getBulkString(str)))

	default:
		c.Write([]byte(getSimpleString("PONG")))
	}
	
}

func getSimpleString(val string) string {
	return fmt.Sprintf("+%s\r\n", val)
}

func getBulkString(val string) string {
	if len(val) == 0 {
		return fmt.Sprintf("$%d\r\n", -1)
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
}