package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

var (
	storedData map[string]string = map[string]string{}
)

func main() {
	fmt.Println("Logs from your program will appear here!")
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
				handleCommand(c, req)
				buffer =  make([]byte, 128)
			}
			
		}()
	}
	
	
}

func handleCommand(c net.Conn, data string) {
	
	if isValid := validateCommand(data); !isValid {
		return
	}
	cmd, err := getCommand(data)

	if err != nil {
		return
	}

	switch *cmd {
	case "ping":
		c.Write([]byte(getSimpleString("PONG")))
	case "echo":
		str, err := handleString(data)
		if err != nil {
			fmt.Println("ERror is ", err.Error())
			c.Write([]byte("+ERROR\r\n"))
		}
		if err == nil {
			c.Write([]byte(fmt.Sprintf("+%s\r\n", *str)))
		}
	case "set":
		if saveData(data) {
			c.Write([]byte(getSimpleString("OK")))
		}else {
			c.Write([]byte(getSimpleString("ERROR")))
		}
	case "get":
		str := getData(data)
		if str != nil {
			c.Write([]byte(getBulkString(*str)))
		} else {
			c.Write([]byte(getBulkString("")))
		}

	default:
		c.Write([]byte("+PING\r\n"))
	}
	
}

func handleString(data string) (*string, error) {
	split := strings.Split(data, "\r\n")
	var ans string = ""
	for i := 3;i<len(split)-1;i+=2 {
		desiredStrLen, err := strconv.Atoi(split[i][1:])
		if err != nil {
			return nil, err
		}
		if desiredStrLen != len(split[i+1]) {
			return nil, errors.New("invalid command")
		}
		ans = fmt.Sprintf("%s %s", ans, split[i+1])
	}
	ans = strings.TrimLeft(ans, " ")
	return &ans, nil
}


func getCommand(data string) (*string, error) {
	split := strings.Split(data, "\r\n")
	if len(split) > 2 {
		return &split[2], nil
	}
	return nil, errors.New("no command")
}

func validateCommand(data string) bool {
	split := strings.Split(data, "\r\n")
	firstLine := split[0]
	firstLineSplit := firstLine[1:]
	noOfWord, err := strconv.Atoi(firstLineSplit)
	if err != nil {
		return false
	}
	if firstLine[0] != '*' || (noOfWord * 2 +2) != len(split) {
		return false
	}
	return true
}

func saveData(data string) bool {
	split := strings.Split(data, "\r\n")
	if len(split) < 8 {
		return false
	}
	storedData[split[4]] = split[6]
	return true
}

func getData(data string) *string {
	split := strings.Split(data, "\r\n")
	if len(split) < 6 {
		return nil
	}
	_data := storedData[split[4]]
	if len(_data) == 0 {
		return nil
	}
	return &_data
	
}

func getSimpleString(val string) string {
	return fmt.Sprintf("+%s\r\n", val)
}

func getBulkString(val string) string {
	if len(val) == 0 {
		return fmt.Sprintf("%d\r\n", -1)
	}
	return fmt.Sprintf("$%d\r\n%s\r\n", len(val), val)
}