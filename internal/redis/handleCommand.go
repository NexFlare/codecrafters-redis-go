package redis

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

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

func HexToBin(hex string) (string, error) {
	ui, err := strconv.ParseUint(hex, 16, 64)
	if err != nil {
		return "", err
	}

	// %016b indicates base 2, zero padded, with 16 characters
	return fmt.Sprintf("%016b", ui), nil
}