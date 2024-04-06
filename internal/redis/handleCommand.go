package redis

import (
	"fmt"
	"strings"
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