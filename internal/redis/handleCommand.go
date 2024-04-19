package redis

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/internal/command"
	"github.com/codecrafters-io/redis-starter-go/internal/response"
)

func handleEchoCommand(args []string) string {
	return strings.Join(args, " ")
}

func handleInfoCommand(rep Replication) string {
	var str string
	str = fmt.Sprintf("role:%s\r\nmaster_repl_offset:%d",rep.Role, rep.MasterReplOffset)
	if len(rep.MasterReplid) > 0 {
		str = fmt.Sprintf("%s\r\nmaster_replid:%s",str, rep.MasterReplid)
	}
	return str
}

func(r *Redis) handlePsyncCommand(f func(string)) {
	if r.Replication.Role == "master" {
		f(response.GetSimpleString(fmt.Sprintf("FULLRESYNC %s 0", r.Replication.MasterReplid)))
		fileContent := "UkVESVMwMDEx+glyZWRpcy12ZXIFNy4yLjD6CnJlZGlzLWJpdHPAQPoFY3RpbWXCbQi8ZfoIdXNlZC1tZW3CsMQQAPoIYW9mLWJhc2XAAP/wbjv+wP9aog=="
		bin, err := base64.StdEncoding.DecodeString(fileContent)
		if err == nil {
			f(response.GetFileString(string(bin)))
		}
	}
}

func(r *Redis) handleSetCommand(cmd *command.Command, f func(string)) {
	var responseString string
	var hasError bool
	var err error
	if len(cmd.Arguments) == 2 {
		r.Store.Set(cmd.Arguments[0], cmd.Arguments[1])
		responseString = response.GetSimpleString("OK")
	} else if len(cmd.Arguments) == 4 {
		duration, err := strconv.Atoi(cmd.Arguments[3])
		if err != nil {
			responseString = response.GetSimpleString("ERROR")
			hasError = true
		} else {
			r.Store.SetWithExpiry(cmd.Arguments[0], cmd.Arguments[1], int64(duration))
			responseString = response.GetSimpleString("OK")
		}
	} else {
		responseString = response.GetSimpleString("")
	}
	if !hasError {
		fmt.Println("The length of replication connection is", len(r.Replication.ReplicationConnection))
		for _, conn := range(r.Replication.ReplicationConnection) {
			regenratedCmd := response.GetBulkString(string(cmd.Command) + " " + strings.Join(cmd.Arguments, " "))
			conn.Write([]byte(fmt.Sprintf("*%d\r\n%s", len(cmd.Arguments) + 1, regenratedCmd)))
		}
	} else {
		fmt.Println("Error while setting value", err.Error())
	}
	f(responseString)
}