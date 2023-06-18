package main

import (
	"fmt"
	"os"

	"github.com/bocha-io/garnet/x/indexer"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/logger"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("ERROR: missing the argument rpc endpoint, ie. http://localhost:8545")
		return
	}

	// Log to file
	file := logger.LogToFile("indexerlogs.txt")
	defer file.Close()

	// Index the database
	quit := false
	database := data.NewDatabase()
	go indexer.Process(os.Args[1], database, &quit)

	// Set up the GUI
	ui := NewDebugUI()
	defer ui.ui.Close()

	// Update each GUI table content
	go ui.ProcessIncomingData(database)
	go ui.ProcessBlockchainInfo(database)
	go ui.ProcessLatestEvents(database)

	// Display the GUI
	ui.Run()

	// Exit program
	quit = true
}
