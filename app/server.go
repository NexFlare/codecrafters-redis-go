package main

import (
	"fmt"
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
		buffer := make([]byte, 128)
		go func() {
			for {
				_, err = c.Read(buffer)
				if err != nil {
					fmt.Println("Error reading command: ", err.Error())
					return
				}
				fmt.Println("Command is: ", string(buffer))
				c.Write([]byte("+PONG\r\n"))
			}
			
		}()
	}
	
	
}
