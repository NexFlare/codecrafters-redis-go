package main

import (
	"fmt"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:6379")
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
				split := strings.Split(req, "\r\n")
				var cmd string
				if len(split) >= 3 {
					cmd = split[2]
				}
				switch cmd {
				case "ping":
					c.Write([]byte("+PONG\r\n"))
				case "echo":
					if len(split) < 5 {
						c.Write([]byte("+invalid command\r\n"))
						break
					}
					message := split[4]
					for i:=6;i<len(split);i+=2 {
						message = fmt.Sprintf("%s %s", message, split[i])
						fmt.Println("Message in loop is ", message)
					}
					fmt.Println("Message is", message)
					c.Write([]byte(fmt.Sprintf("+%s\r\n",message)))
				default:
					c.Write([]byte("+PONG\r\n"))

				}
				buffer =  make([]byte, 128)
			}
			
		}()
	}
	
	
}
