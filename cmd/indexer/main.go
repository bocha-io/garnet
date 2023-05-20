package main

import (
	"log"
	"os"

	"github.com/hanchon/garnet/internal/backend"
	"github.com/hanchon/garnet/internal/indexer"
	"github.com/hanchon/garnet/internal/indexer/data"
)

const (
	port = 6666
)

func main() {
	// Set the log output to a file (stdin, stdout, stderror used by GUI)
	fileName := "logFile.log"
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)

	// Index the database
	quit := false
	database := data.NewDatabase()
	go indexer.Process(&database, &quit)

	// Set up the GUI
	ui := NewDebugUI()
	defer ui.ui.Close()

	go ui.ProcessIncomingData(&database)
	go ui.ProcessBlockchainInfo(&database)
	go ui.ProcessLatestEvents(&database)

	// Start the backend server
	go backend.StartGorillaServer(port)

	// Display the GUI
	ui.Run()

	// Exit program
	quit = true
}
