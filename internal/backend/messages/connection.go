package messages

import (
	"fmt"
	"net/http"
	"sync"
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
	Spectators        *map[string]*[]string
	spectatorsMutex   *sync.Mutex
	UsersDatabase     *database.InMemoryDatabase
	Database          *data.Database
	LastBroadcastTime time.Time
}

func (g *GlobalState) addSpectator(gameID, userID string) {
	g.spectatorsMutex.Lock()
	defer g.spectatorsMutex.Unlock()
	if v, ok := (*g.Spectators)[gameID]; ok {
		found := false
		for _, user := range *v {
			if user == userID {
				found = true
				break
			}
		}
		if !found {
			temp := append(*v, userID)
			(*g.Spectators)[gameID] = &temp
		}
	} else {
		(*g.Spectators)[gameID] = &[]string{userID}
	}
}

func (g *GlobalState) rmSpectator(gameID, userID string) {
	g.spectatorsMutex.Lock()
	defer g.spectatorsMutex.Unlock()
	if v, ok := (*g.Spectators)[gameID]; ok {
		found := -1
		for k, user := range *v {
			if user == userID {
				found = k
				break
			}
		}

		if found != -1 {
			var temp []string
			if len(*v)-1 <= found {
				temp = (*v)[0:found]
			} else {
				temp = append((*v)[0:found], (*v)[found+1:len(*v)-1]...)
			}
			(*g.Spectators)[gameID] = &temp
			return
		}
	}
}

func (g *GlobalState) boardcastToSpectators(gameID string, response interface{}) {
	spectators, ok := (*g.Spectators)[gameID]
	if !ok {
		return
	}

	for _, v := range *spectators {
		if ws, ok := g.WsSockets[v]; ok {
			err := ws.Conn.WriteJSON(response)
			if err != nil {
				logger.LogError(fmt.Sprintf("[backend] error sending spectator update to %s", v))
			} else {
				logger.LogDebug(fmt.Sprintf("[backend] sending spectator update to %s", v))
			}
		}
	}
}

func NewGlobalState(database *data.Database, usersDatabase *database.InMemoryDatabase) GlobalState {
	temp := make(map[string]*[]string)
	return GlobalState{
		done:              make(chan struct{}),
		WalletIndex:       make(map[string]string),
		WsSockets:         make(map[string]*WebSocketContainer),
		Spectators:        &temp,
		spectatorsMutex:   &sync.Mutex{},
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

										// Player Two logic
										playerTwoValue := []byte{}
										_, playerTwo, err := actions.GetPlayerTwoFromGame(g.Database, w, k)
										if err == nil {
											_, playerTwoName, err := actions.GetUserName(g.Database, w, playerTwo)
											if err != nil {
												logger.LogError("[backend] could not find player name")
												continue
											}
											playerTwoValue, err = hexutil.Decode(playerTwoName)
											if err != nil {
												logger.LogError("[backend] could not decode players name")
												continue
											}
										}

										// PlayerOne values
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

										ret = append(ret, Match{Id: k, Creator: string(temp), PlayerTwo: string(playerTwoValue)})
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
