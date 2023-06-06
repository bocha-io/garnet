package actions

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hanchon/garnet/internal/backend/messages/dbconnector"
	"github.com/hanchon/garnet/internal/indexer/data"
	"github.com/hanchon/garnet/internal/logger"
	"github.com/hanchon/garnet/internal/txbuilder"
)

func validateMove(db *data.Database, cardID [32]byte, msg *MoveCard, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)

	gameKey, err := commonValidation(db, w, cardID, walletAddress, moveManaCost)
	if err != nil {
		return false, err
	}

	_, err = GetCardInPosition(db, w, gameKey, msg.X, msg.Y)
	// The space must be empty
	if err == nil {
		logger.LogError(fmt.Sprintf("[backend] there is a unit at position (%d, %d) game %s", msg.X, msg.Y, gameKey))
		return false, nil
	}

	actionReady := IsCardReady(db, w, hexutil.Encode(cardID[:]))
	if !actionReady {
		logger.LogError(fmt.Sprintf("[backend] card already attacked: %s", hexutil.Encode(cardID[:])))
		return false, nil
	}

	// Range
	position, err := GetCardPosition(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError("[backend] could not get the card position")
		return false, err
	}

	if position.X < 0 || position.X > 9 {
		logger.LogError("[backend] the card was not placed")
		return false, nil
	}

	movementSpeed, err := GetCardMovementSpeed(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError("[backend] could not get the card mov speed")
		return false, err
	}

	deltaX := math.Abs(float64(position.X - msg.X))
	deltaY := math.Abs(float64(position.Y - msg.Y))
	if deltaX+deltaY > float64(movementSpeed) {
		logger.LogError("[backend] trying to move too far away")
		return false, nil
	}

	return true, nil
}

func movePrediction(db *data.Database, cardID [32]byte, msg *MoveCard, txhash common.Hash) (string, MoveCardResponse, error) {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return "", MoveCardResponse{}, err
	}

	_, cardOwner, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the card owner %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", MoveCardResponse{}, err
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return "", MoveCardResponse{}, err
	}

	events := []data.MudEvent{
		{
			Table: "Position",
			Key:   hexutil.Encode(cardID[:]),
			Fields: []data.Field{
				{Key: "placed", Data: data.BoolField{Data: true}},
				{Key: "gamekey", Data: gameField},
				{Key: "x", Data: data.UintField{Data: *big.NewInt(msg.X)}},
				{Key: "y", Data: data.UintField{Data: *big.NewInt(msg.Y)}},
			},
		},
		{
			Table: "CurrentMana",
			Key:   gameKey,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - moveManaCost)}},
			},
		},
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: events,
	},
	)
	response := MoveCardResponse{
		UUID:         msg.UUID,
		MsgType:      "movecardresponse",
		CardID:       msg.CardID,
		EndX:         msg.X,
		EndY:         msg.Y,
		Player:       cardOwner,
		LeftOverMana: currentMana - moveManaCost,
	}

	return gameKey, response, nil
}

func MoveHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, MoveCardResponse, error) {
	// TODO: Wallet address is used to validate the action
	_ = walletAddress
	if !authenticated {
		return "", MoveCardResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing move card request")

	var msg MoveCard
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding move card message: %s", err))
		return "", MoveCardResponse{}, nil
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to move card: %s", err))
		return "", MoveCardResponse{}, nil
	}

	valid, err := validateMove(db, cardID, &msg, walletAddress)
	if err != nil || !valid {
		return "", MoveCardResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "movecard", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to move card: %s", err))
		return "", MoveCardResponse{}, nil
	}

	gameID, response, err := movePrediction(db, cardID, &msg, txhash)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction move: %s", err))
		return "", MoveCardResponse{}, err
	}
	return gameID, response, nil
}
