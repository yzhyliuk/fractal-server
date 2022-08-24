package logs

import (
	"fmt"
	"time"
)

var Debug = true

func LogDebug(str string, err error) {
	flag := "DEBUG |"
	message := ""
	if err != nil {
		flag = "ERROR |"
		message = err.Error()
	}
	if Debug {
		fmt.Println(time.Now().Format(time.RFC1123), flag, str, message)
	}
}

func LogError(err error)  {
	flag := ""
	message := ""
	if err != nil {
		flag = "ERROR |"
		message = err.Error()
	} else {
		return
	}
	fmt.Println(time.Now().Format(time.RFC1123), flag, message)
}