package command

type CommandType string

const (
	PING CommandType = "ping"
	SET CommandType = "set"
	GET CommandType = "get"
	ECHO CommandType = "echo"
)