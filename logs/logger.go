package logs

import (
	"bufio"
	"fmt"
	"log"
	"os"
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

func LogError(err error) {
	flag := ""
	message := ""
	if err != nil {
		flag = "ERROR |"
		message = err.Error()
	} else {
		return
	}
	fmt.Println(time.Now().Format(time.RFC1123), flag, message)

	file, err := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	datawriter := bufio.NewWriter(file)

	_, _ = datawriter.WriteString(time.Now().Format(time.RFC1123) + "| ERROR |" + message + "\n")
	datawriter.Flush()
	file.Close()
}
