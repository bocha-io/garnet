package messages

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bocha-io/garnet/internal/backend/messages/actions"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/bocha-io/garnet/internal/txbuilder"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
)

func writeMessage(ws *websocket.Conn, msg *string) error {
	return ws.WriteMessage(websocket.TextMessage, []byte(*msg))
}

func removeConnection(ws *WebSocketContainer, g *GlobalState) {
	ws.Conn.Close()
	delete(g.WsSockets, ws.User)
}

func (g *GlobalState) WsHandler(ws *WebSocketContainer) {
	for {
		defer removeConnection(ws, g)
		// Read until error the client messages
		_, p, err := ws.Conn.ReadMessage()
		if err != nil {
			return
		}

		// TODO: log ip address
		logger.LogDebug(fmt.Sprintf("[backend] incoming message: %s", string(p)))

		var m BasicMessage
		err = json.Unmarshal(p, &m)
		if err != nil {
			return
		}

		switch m.MsgType {
		case "connect":
			if connectMessage(ws, g.UsersDatabase, &p) != nil {
				return
			}

			// Send response
			msg := `{"msgtype":"connected", "status":true}`
			if writeMessage(ws.Conn, &msg) != nil {
				return
			}

			logger.LogDebug(fmt.Sprintf("[backend] senging message: %s", msg))

			g.WsSockets[ws.User] = ws

			w, ok := g.Database.Worlds[actions.WorldID]
			if !ok {
				panic("world not found")
			}

			if err := ws.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"msgtype":"userwallet","wallet":"%s"}`, ws.WalletAddress))); err != nil {
				panic("could not send the board status")
			}

			matchData := actions.GetBoardStatus(g.Database, actions.WorldID, ws.WalletAddress)
			if matchData != nil {
				logger.LogInfo(fmt.Sprintf("[test] names %s %s", matchData.PlayerOneUsermane, matchData.PlayerTwoUsermane))
				msgToSend := BoardStatus{MsgType: "boardstatus", Status: *matchData}
				logger.LogDebug(fmt.Sprintf("[backend] sending match info %s to %s", matchData.MatchID, ws.User))
				err := ws.Conn.WriteJSON(msgToSend)
				if err != nil {
					panic("could not send the board status")
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
					err := ws.Conn.WriteJSON(msg)
					if err != nil {
						// TODO: close the connection
						return
					}
					logger.LogDebug(fmt.Sprintf("[backend] sending %d active matches", len(ret)))
				}
			}

		case "creatematch":
			if !ws.Authenticated {
				return
			}
			_, err := txbuilder.SendTransaction(ws.WalletID, "creatematch")
			if err != nil {
				// TODO: send response saying that the game could not be created
				logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to creatematch: %s", err))
			}
			// g.Database.AddTxSent(txhash.Hex())

		case "joinmatch":
			if !ws.Authenticated {
				return
			}
			logger.LogDebug("[backend] processing join match request")

			var msg JoinMatch
			err := json.Unmarshal(p, &msg)
			if err != nil {
				logger.LogError(fmt.Sprintf("[backend] error decoding join match message: %s", err))
				// return
			}

			logger.LogDebug(fmt.Sprintf("[backend] creating join match tx: %s", msg.MatchID))

			id, err := hexutil.Decode(msg.MatchID)
			if err != nil {
				// TODO: send response saying that the game could not be created
				logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to join match: %s", err))
				// return
			}

			if len(id) != 32 {
				logger.LogDebug("[backend] error creating transaction to join match: invalid length")
				// return
			}

			// It must be array instead of slice
			var idArray [32]byte
			copy(idArray[:], id)

			_, err = txbuilder.SendTransaction(ws.WalletID, "joinmatch", idArray)
			if err != nil {
				// TODO: send response saying that the game could not be created
				logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to join match: %s", err))
				// return
			}
			// g.Database.AddTxSent(txhash.Hex())

		case "endturn":
			gameID, response, err := actions.EndturnHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "surrender":
			gameID, response, err := actions.SurrenderHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "movecard":
			gameID, response, err := actions.MoveHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()
		case "attack":
			gameID, response, err := actions.AttackHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()
		case "placecard":
			gameID, response, err := actions.PlaceCardHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "meteor":
			gameID, response, err := actions.MeteorHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "drainsword":
			gameID, response, err := actions.DrainSwordHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "piercingshot":
			gameID, response, err := actions.PiercingShotHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "whirlwindaxe":
			gameID, response, err := actions.WhirldwindAxeHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "sidestep":
			gameID, response, err := actions.SidestepHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		case "cover":
			gameID, response, err := actions.CoverHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			if err != nil {
				logger.LogError(fmt.Sprintf("[test] cover failed %s", err.Error()))
				return
			}
			broadcastResponse(g, gameID, response)
			g.Database.LastUpdate = time.Now()

		}
	}
}

func broadcastResponse(g *GlobalState, gameID string, response interface{}) {
	w := g.Database.GetWorld(actions.WorldID)

	logger.LogInfo(fmt.Sprintf("[test] trying to send to player in game %s", gameID))

	_, playerOne, err := actions.GetPlayerOneFromGame(g.Database, w, gameID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not find player one in game %s", gameID))
		return
	}

	_, playerTwo, err := actions.GetPlayerTwoFromGame(g.Database, w, gameID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not find player two in game %s", gameID))
		return
	}

	for _, v := range g.WsSockets {
		if strings.Contains(strings.ToLower(playerOne), v.WalletAddress[2:]) {
			err := v.Conn.WriteJSON(response)
			if err != nil {
				logger.LogError(fmt.Sprintf("[backend] error sending update to %s", v.WalletAddress))
			} else {
				logger.LogDebug(fmt.Sprintf("[backend] sending update to %s", v.WalletAddress))
			}
		}
		if strings.Contains(strings.ToLower(playerTwo), v.WalletAddress[2:]) {
			err := v.Conn.WriteJSON(response)
			if err != nil {
				logger.LogError(fmt.Sprintf("[backend] error sending update to %s", v.WalletAddress))
			} else {
				logger.LogDebug(fmt.Sprintf("[backend] sending update to %s", v.WalletAddress))
			}
		}
	}
}
