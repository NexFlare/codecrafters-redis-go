package response

import (
	"fmt"
	"strconv"
	"strings"
)

func GetSimpleString(val string) string {
	return fmt.Sprintf("+%s\r\n", val)
}

func GetBulkString(val string) string {
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

func GetArrayString(val string) string {
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

func GetFileString(val string) string {
	if len(val) == 0 {
		return fmt.Sprintf("$%d\r\n", -1)
	}

	str := GetBulkString(val)
	return str[:len(str)-2]
}