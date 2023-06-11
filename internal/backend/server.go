package backend

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bocha-io/garnet/internal/backend/api"
	"github.com/bocha-io/garnet/internal/backend/cors"
	"github.com/bocha-io/garnet/internal/backend/messages"
	"github.com/bocha-io/garnet/internal/database"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/gorilla/mux"
)

func StartGorillaServer(port int, mudDatabase *data.Database) error {
	logger.LogInfo(fmt.Sprintf("[backend] starting server at port: %d\n", port))
	router := mux.NewRouter()
	usersDatabase := database.NewInMemoryDatabase()
	g := messages.NewGlobalState(mudDatabase, usersDatabase)
	router.HandleFunc("/ws", g.WebSocketConnectionHandler).Methods("GET", "OPTIONS")
	api.RestRoutes(router, usersDatabase)
	go g.BroadcastUpdates()

	cors.ServerEnableCORS(router)

	server := &http.Server{
		Addr:              fmt.Sprint(":", port),
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}
	return server.ListenAndServe()
}
