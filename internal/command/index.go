package command

import (
	"fmt"
	"strconv"
	"strings"
)

type Command struct {
	Command CommandType
	Arguments []string
}

func NewCommand(data string) *Command {
	cmd, args, err := parseCommand(data)
	if err != nil {
		fmt.Println("Error while creating command", err.Error())
		return nil
	}
	return &Command{
		Command: CommandType(cmd),
		Arguments: args,
	}
}

// CMD is *3
// $3
// set
// $5
// class
// $5
// cloud

// Clients send commands to a Redis server as an array of bulk strings. 
func parseCommand(data string) (string, []string, error){
	split := strings.Split(data, "\r\n")
	arrSize, err := strconv.Atoi(split[0][1:])
	var cmd string = ""
	var args []string = []string{}
	if err != nil {
		return "", nil, err
	}
	if len(split) > 2 {
		cmd = strings.ToLower(split[2])
	}
	for i := 4; i<=arrSize*2; i+=2  {
		args = append(args, split[i])
	}
	return cmd, args, nil
} 