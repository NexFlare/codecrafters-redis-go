package redis

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/store"
)

type Replication struct {
	Role string
	MasterReplid string
	MasterReplOffset int
	ReplBacklogSize int
	ReplBacklogActive bool
	MasterHost string
	MasterPort string
}


type Redis struct {
	Flags map[string]string
	port int
	Replication Replication
	Store store.Store
}

func NewRedisServer() Redis {
	return Redis{
		port : 6379,
		Replication: Replication{
			Role: "master",
			MasterReplid: "8371b4fb1155b71f4a04d3e1bc3e18c4a990aeeb",
			MasterReplOffset: 0,
		},
		Store: store.GetStore(),
	}
}

func(r* Redis) StartRedis() {
	r.parseFlags()
	fmt.Println("Port is", r.port)
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", r.port))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	fmt.Println("Role is", r.Replication.Role)
	if r.Replication.Role == "slave" {
		go r.handleHandShake()
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
				r.handleCommand(c, req)
				buffer =  make([]byte, 128)
			}
			
		}()
	}
}

func(r* Redis) parseFlags() {
	for i:=0;i<len(os.Args);i++ {
		switch os.Args[i] {
		case "--port":
			i++
			port, err := strconv.Atoi(os.Args[i])
			if err == nil {
				r.port = port
			} else {
				fmt.Println("Error with port", err.Error())
			}
		case "--replicaof":
			r.Replication.Role = "slave"
			i++
			r.Replication.MasterHost = os.Args[i]
			i++
			if strings.Contains(os.Args[i], "--") {
				continue
			}
			r.Replication.MasterPort = os.Args[i]
		}
	}
}

func(r *Redis) handleCommand(c net.Conn, data string) {
	
	cmd := command.NewCommand(data)
	if cmd == nil {
		c.Write([]byte(getSimpleString("ERROR")))
		return
	}
	switch cmd.Command {
	case command.PING:
		c.Write([]byte(getSimpleString("PONG")))
	case command.ECHO:
		c.Write([]byte(getSimpleString(handleEchoCommand(cmd.Arguments))))
	case command.SET:
		if len(cmd.Arguments) == 2 {
			r.Store.Set(cmd.Arguments[0], cmd.Arguments[1])
			c.Write([]byte(getSimpleString("OK")))
		} else if len(cmd.Arguments) == 4 {
			duration, err := strconv.Atoi(cmd.Arguments[3])
			if err != nil {
				c.Write([]byte(getSimpleString("ERROR")))
				return
			}
			r.Store.SetWithExpiry(cmd.Arguments[0], cmd.Arguments[1], int64(duration))
			c.Write([]byte(getSimpleString("OK")))
		} else {
			c.Write([]byte(getSimpleString("")))
		}
	case command.GET:
		str := r.Store.Get(cmd.Arguments[0])
		c.Write([]byte(getBulkString(str)))
	case command.INFO:
		c.Write([]byte(getBulkString(handleInfoCommand(r.Replication))))
		return
	default:
		c.Write([]byte(getSimpleString("PIONG")))
	}
	
}

func getSimpleString(val string) string {
	return fmt.Sprintf("+%s\r\n", val)
}

func getBulkString(val string) string {
	if len(val) == 0 {
		return fmt.Sprintf("$%d\r\n", -1)
	}
	var str string
	split := strings.Split(val, " ")
	for _, item := range(split) {
		str = fmt.Sprintf("%s$%d\r\n%s\r\n", str, len(item), item)
	}
	return str
}

func getArrayString(val string) string {
	splitArr := strings.Split(val, "\r\n")
	totalLen:=0
	for i:=0;i<len(splitArr); {
		if len(splitArr[i]) > 0 {
			totalLen++
			switch splitArr[i][0] {
			case '$':
				i+=2
			case '*':
				nestedArrayLen, err := strconv.Atoi(splitArr[i][1:])
				if err != nil {
					fmt.Println("error while parsing command", err.Error())
				} else {
					i+=(nestedArrayLen+1)
				}
			default:
				i++
			}
		} else {
			i++
		}
 	}
	return fmt.Sprintf("*%d\r\n%s", totalLen, val)
}

func(r* Redis) handleHandShake() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", r.Replication.MasterHost, r.Replication.MasterPort))
	defer conn.Close()
	if err != nil {
		fmt.Println("error is ", err.Error())
		return
	}
	ch := make(chan(int))
	req := getArrayString(getBulkString("ping"))
	go r.handleHandShakeRequest(conn, req, ch)
	resp := <- ch
	if resp == 0 {
		return
	}

	req = getArrayString(getBulkString(fmt.Sprintf("REPLCONF listening-port %d", r.port)))
	go r.handleHandShakeRequest(conn, req, ch)
	resp = <- ch
	if resp == 0 {
		return
	}

	req = getArrayString(getBulkString("REPLCONF capa psync2"))
	go r.handleHandShakeRequest(conn, req, ch)
	resp = <- ch
	if resp == 0 {
		return
	}
}

func(r* Redis) handleHandShakeRequest(conn net.Conn, val string, ch chan int) {
	if len(val) == 0 {
		fmt.Println("NO value of string")
		ch <- 0
	}
	_, err := conn.Write([]byte(val))

	if err != nil {
		fmt.Println("error is ", err.Error())
		ch <- 0
		return
	}
	reply := make([]byte, 256)
	_, err = conn.Read(reply)
	fmt.Println("Reply is ", string(reply))
	if err != nil {
		fmt.Println("error is ", err.Error())
		ch <- 0
		return
	}
	ch <- 1

}