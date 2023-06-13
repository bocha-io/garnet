package messages

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bocha-io/garnet/internal/backend/cors"
	"github.com/bocha-io/garnet/internal/backend/messages/actions"
	"github.com/bocha-io/garnet/internal/database"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type WebSocketContainer struct {
	Authenticated bool
	User          string
	WalletID      int
	WalletAddress string
	Conn          *websocket.Conn
}

type User struct {
	Username string
	Password string
	WalletID int
}

type GlobalState struct {
	done              chan (struct{})
	WalletIndex       map[string]string
	WsSockets         map[string]*WebSocketContainer
	UsersDatabase     *database.InMemoryDatabase
	Database          *data.Database
	LastBroadcastTime time.Time
}

func NewGlobalState(database *data.Database, usersDatabase *database.InMemoryDatabase) GlobalState {
	return GlobalState{
		done:              make(chan struct{}),
		WalletIndex:       make(map[string]string),
		WsSockets:         make(map[string]*WebSocketContainer),
		UsersDatabase:     usersDatabase,
		Database:          database,
		LastBroadcastTime: time.Now(),
	}
}

func (g *GlobalState) WebSocketConnectionHandler(response http.ResponseWriter, request *http.Request) {
	if cors.SetHandlerCorsForOptions(request, &response) {
		return
	}

	// TODO: Filter prod page or localhost for development
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		// Maybe log the error
		return
	}

	webSocket := WebSocketContainer{
		Authenticated: false,
		Conn:          ws,
	}

	g.WsHandler(&webSocket)
}

func (g *GlobalState) BroadcastUpdates() {
	for {
		select {
		case <-g.done:
			return
		case <-time.After(50 * time.Millisecond):
			if len(g.WsSockets) != 0 {
				timestamp := g.Database.LastUpdate
				if g.LastBroadcastTime != timestamp {
					logger.LogDebug("[backend] database was updated, broadcasting messages...")

					w, ok := g.Database.Worlds[actions.WorldID]
					if !ok {
						panic("world not found")
					}

					g.LastBroadcastTime = timestamp

					for _, v := range g.WsSockets {
						if v.Conn != nil {
							matchData := actions.GetBoardStatus(g.Database, actions.WorldID, v.WalletAddress)
							if matchData != nil {
								logger.LogInfo(fmt.Sprintf("[test] names %s %s", matchData.PlayerOneUsermane, matchData.PlayerTwoUsermane))
								msgToSend := BoardStatus{MsgType: "boardstatus", Status: *matchData}
								logger.LogDebug(fmt.Sprintf("[backend] sending match info %s to %s", matchData.MatchID, v.User))
								err := v.Conn.WriteJSON(msgToSend)
								if err != nil {
									logger.LogError(fmt.Sprintf("[backend] error sending transaction to client, unsubscribing: %s", err))
									// TODO: unsub from this connection, it requires a lock in the g.WsSockets variable to avoid breaking the loops
									v.Authenticated = false
									v.Conn = nil
								}
							} else {
								t := w.GetTableByName("Match")
								if t != nil {
									ret := []Match{}
									for k := range *t.Rows {
										_, playerOne, err := actions.GetPlayerOneFromGame(g.Database, w, k)
										if err != nil {
											continue
										}

										_, _, err = actions.GetPlayerTwoFromGame(g.Database, w, k)
										if err == nil {
											// Ignore games if it has 2 players
											continue
										}

										_, playerOneName, err := actions.GetUserName(g.Database, w, playerOne)
										if err != nil {
											logger.LogError("[backend] match does not have a player one")
											continue
										}
										temp, err := hexutil.Decode(playerOneName)
										if err != nil {
											logger.LogError("[backend] could not decode players name")
											continue
										}

										ret = append(ret, Match{Id: k, Creator: string(temp)})
									}
									msg := MatchList{MsgType: "matchlist", Matches: ret}
									err := v.Conn.WriteJSON(msg)
									if err != nil {
										logger.LogError(fmt.Sprintf("[backend] error sending transaction to client, unsubscribing: %s", err))
										// TODO: unsub from this connection, it requires a lock in the g.WsSockets variable to avoid breaking the loops
										v.Authenticated = false
										v.Conn = nil
									}
									logger.LogDebug(fmt.Sprintf("[backend] sending %d active matches", len(ret)))
								}
							}
						}
					}
				}
			}
		}
	}
}
