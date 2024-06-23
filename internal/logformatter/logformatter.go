package logformatter

import (
	"log"
	"strings"
	"time"
)

func LogError(errStrings []string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	log.Println(now)
	log.Fatal(strings.Join(errStrings, "\n"))
}
