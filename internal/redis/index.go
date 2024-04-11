package redis

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/response"
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
	ReplicationConnection []net.Conn
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
			ReplicationConnection: []net.Conn{}, 
		},
		Store: store.GetStore(),
	}
}

func(r* Redis) StartRedis() {
	r.parseFlags()
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
		defer c.Close()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go func() {
			buffer := make([]byte, 128)
			for {
				_, err = c.Read(buffer)
				if err != nil {
					fmt.Println("Error reading command: ", err.Error())
					return
				}
				req := string(buffer)
				r.handleCommand(c, req, true)
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

func(r *Redis) handleCommand(c net.Conn, data string, sendResponse bool) {
	
	cmd := command.NewCommand(data)
	if cmd == nil {
		// c.Write([]byte(response.GetSimpleString("ERROR")))
		return
	}
	switch cmd.Command {
	case command.PING:
		c.Write([]byte(response.GetSimpleString("PONG")))
	case command.ECHO:
		c.Write([]byte(response.GetSimpleString(handleEchoCommand(cmd.Arguments))))
	case command.SET:
		r.handleSetCommand(cmd, func(s string) {
			if sendResponse {
				c.Write([]byte(s))
			}
		})
	case command.GET:
		str := r.Store.Get(cmd.Arguments[0])
		c.Write([]byte(response.GetBulkString(str)))
	case command.INFO:
		c.Write([]byte(response.GetBulkString(handleInfoCommand(r.Replication))))
		return
	case command.PSYNC:
		r.Replication.ReplicationConnection = append(r.Replication.ReplicationConnection, c)
		r.handlePsyncCommand(func (s string) {
			c.Write([]byte(s))
		})
		return
	default:
		c.Write([]byte(response.GetSimpleString("OK")))
	}
	
}

func(r* Redis) handleHandShake() {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", r.Replication.MasterHost, r.Replication.MasterPort))
	// Not closing the connection as it will be used for replication
	// defer conn.Close()
	if err != nil {
		fmt.Println("error is ", err.Error())
		return
	}
	ch := make(chan(int))

	handShakeArray := []string {
		"ping",
		fmt.Sprintf("REPLCONF listening-port %d", r.port),
		"REPLCONF capa psync2",
		"PSYNC ? -1",
	}

	for _, s := range(handShakeArray) {
		s = response.GetArrayString(response.GetBulkString(s))
		go r.handleHandShakeRequest(conn, s, ch)
		resp := <- ch
		if resp == 0 {
			fmt.Println("Error in handshake")
			conn.Close()
			return
		}
	}
	
	for {
		buffer := make([]byte, 128)
		_, err = conn.Read(buffer)
		if err == nil {
			// fmt.Println("Got command ", string(buffer))
			strBuffer := strings.Trim(string(buffer), "\x00")
			if len(strBuffer) > 0 {
				fmt.Println("Got command ", strBuffer)
				r.handleCommand(conn, strBuffer, false)
			}
		}
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
	if err != nil {
		fmt.Println("error is ", err.Error())
		ch <- 0
		return
	}
	ch <- 1

}