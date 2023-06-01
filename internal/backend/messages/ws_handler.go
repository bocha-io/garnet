package messages

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gorilla/websocket"
	"github.com/hanchon/garnet/internal/backend/messages/actions"
	"github.com/hanchon/garnet/internal/logger"
	"github.com/hanchon/garnet/internal/txbuilder"
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

			matchData := actions.GetBoardStatus(g.Database, actions.WorldID, ws.WalletAddress)
			if matchData != nil {
				// TODO: save the user and wallet somewhere
				matchData.PlayerOneUsermane = "user1"
				matchData.PlayerTwoUsermane = "user2"
				msgToSend := BoardStatus{MsgType: "boardstatus", Status: *matchData}
				logger.LogDebug(fmt.Sprintf("[backend] sending match info %s to %s", matchData.MatchID, ws.User))
				err := ws.Conn.WriteJSON(msgToSend)
				if err != nil {
					panic("could not send the board status")
				}
			} else {
				t := w.GetTableByName("Match")
				if t != nil {
					ret := []string{}
					for k := range *t.Rows {
						ret = append(ret, k)
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
				return
			}

			logger.LogDebug(fmt.Sprintf("[backend] creating join match tx: %s", msg.MatchID))

			id, err := hexutil.Decode(msg.MatchID)
			if err != nil {
				// TODO: send response saying that the game could not be created
				logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to join match: %s", err))
				return
			}

			if len(id) != 32 {
				logger.LogDebug("[backend] error creating transaction to join match: invalid length")
				return
			}

			// It must be array instead of slice
			var idArray [32]byte
			copy(idArray[:], id)

			_, err = txbuilder.SendTransaction(ws.WalletID, "joinmatch", idArray)
			if err != nil {
				// TODO: send response saying that the game could not be created
				logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to join match: %s", err))
				return
			}
			// g.Database.AddTxSent(txhash.Hex())

		case "endturn":
			actions.EndturnHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			g.Database.LastUpdate = time.Now()
		case "movecard":
			actions.MoveHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			g.Database.LastUpdate = time.Now()
		case "attack":
			actions.AttackHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			g.Database.LastUpdate = time.Now()
		case "placecard":
			actions.PlaceCardHandler(ws.Authenticated, ws.WalletID, ws.WalletAddress, g.Database, p)
			g.Database.LastUpdate = time.Now()
		}
	}
}
