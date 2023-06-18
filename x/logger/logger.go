package logger

import (
	"log"
	"os"
)

func LogInfo(value string) {
	log.Printf("[INFO]:  %s", value)
}

func LogDebug(value string) {
	log.Printf("[DEBUG]: %s", value)
}

func LogError(value string) {
	log.Printf("[ERROR]: %s", value)
}

func LogWarning(value string) {
	log.Printf("[WARN]:  %s", value)
}

// Call `defer logFile.Close()` after this function
func LogToFile(fileName string) *os.File {
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		log.Panic(err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)
	return logFile
}
