package actions

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hanchon/garnet/internal/backend/messages/dbconnector"
	"github.com/hanchon/garnet/internal/indexer/data"
	"github.com/hanchon/garnet/internal/logger"
	"github.com/hanchon/garnet/internal/txbuilder"
)

func endturnPrediction(db *data.Database, gameID [32]byte, txhash common.Hash, msgUUID string) (string, EndTurnResponse, error) {
	w := db.GetWorld(WorldID)
	gameIDAsString := hexutil.Encode(gameID[:])

	_, currentPlayer, err := GetCurrentPlayerFromGame(db, w, gameIDAsString)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error getting current player  %s, %s", gameIDAsString, err.Error()))
		return "", EndTurnResponse{}, err
	}

	playerOneField, playerOne, err := GetPlayerOneFromGame(db, w, gameIDAsString)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error getting player one  %s, %s", gameIDAsString, err.Error()))
		return "", EndTurnResponse{}, err
	}

	playerTwoField, playerTwo, err := GetPlayerTwoFromGame(db, w, gameIDAsString)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error getting player two  %s, %s", gameIDAsString, err.Error()))
		return "", EndTurnResponse{}, err
	}

	currentTurn, err := GetCurrentTurnFromGame(db, w, gameIDAsString)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error getting current turn  %s, %s", gameIDAsString, err.Error()))
		return "", EndTurnResponse{}, err
	}

	newPlayer := playerOneField
	newPlayerString := playerOne
	if currentPlayer == playerOne {
		newPlayer = playerTwoField
		newPlayerString = playerTwo
	}

	newMana := currentTurn + 5 + 1
	if newMana > 15 {
		newMana = 15
	}

	events := []data.MudEvent{
		{
			Table: "CurrentPlayer",
			Key:   gameIDAsString,
			Fields: []data.Field{
				newPlayer,
			},
		},
		{
			Table: "CurrentMana",
			Key:   gameIDAsString,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(newMana)}},
			},
		},
		{
			Table: "CurrentTurn",
			Key:   gameIDAsString,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentTurn + 1)}},
			},
		},
	}

	cards := GetCardsFromMatch(db, w, gameIDAsString)
	for _, v := range cards {
		events = append(events, data.MudEvent{
			Table: "ActionReady",
			Key:   v,
			Fields: []data.Field{
				{Key: "value", Data: data.BoolField{Data: true}},
			},
		})
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: events,
	},
	)

	return gameIDAsString, EndTurnResponse{UUID: msgUUID, MsgType: "endturnresponse", Player: newPlayerString, Mana: newMana, Turn: currentTurn + 1}, nil
}

func EndturnHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, EndTurnResponse, error) {
	// TODO: Wallet address is used to validate the action
	_ = walletAddress
	if !authenticated {
		return "", EndTurnResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing endturn request")

	var msg EndTurn
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding endturn message: %s", err))
		return "", EndTurnResponse{}, err
	}

	logger.LogDebug(fmt.Sprintf("[backend] creating endturn tx: %s", msg.MatchID))

	matchID, err := dbconnector.StringToSlice(msg.MatchID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to end turn: %s", err))
		return "", EndTurnResponse{}, err
	}

	txhash, err := txbuilder.SendTransaction(walletID, "endturn", matchID)
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to endturn: %s", err))
		return "", EndTurnResponse{}, err
	}

	gameID, response, err := endturnPrediction(db, matchID, txhash, msg.UUID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction end turn: %s", err))
		return "", EndTurnResponse{}, err
	}
	return gameID, response, nil
}
