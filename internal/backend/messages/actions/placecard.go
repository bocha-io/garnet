package actions

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hanchon/garnet/internal/backend/messages/dbconnector"
	"github.com/hanchon/garnet/internal/indexer/data"
	"github.com/hanchon/garnet/internal/logger"
	"github.com/hanchon/garnet/internal/txbuilder"
)

func validatePlaceCard(db *data.Database, cardID [32]byte, msg *PlaceCard, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)

	gameKey, err := commonValidation(db, w, cardID, walletAddress, summonManaCost)
	if err != nil {
		return false, err
	}

	_, err = GetCardInPosition(db, w, gameKey, msg.X, msg.Y)
	// The space must be empty
	if err == nil {
		logger.LogError(fmt.Sprintf("[backend] there is a unit at position (%d, %d) game %s", msg.X, msg.Y, gameKey))
		return false, fmt.Errorf("invalid place position, there is a unit there")
	}

	position, err := GetCardPosition(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError("[backend] could not get the card position")
		return false, err
	}

	if position.X != -2 || position.Y != -2 {
		logger.LogError("[backend] card was already placed")
		return false, err
	}

	_, playerOne, err := GetPlayerOneFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError("[backend] could not get the player one")
		return false, err
	}

	var minY int64 = 8
	var maxY int64 = 9
	if strings.Contains(playerOne, walletAddress) {
		minY = 0
		maxY = 1
	}

	if msg.X < 0 || msg.X > 9 || msg.Y < minY || msg.Y > maxY {
		logger.LogError("[backend] invalid place position")
		return false, fmt.Errorf("invalid place position")
	}

	return true, nil
}

func placeCardPrediction(db *data.Database, cardID [32]byte, msg *PlaceCard, txhash common.Hash, walletAddress string) (string, PlaceCardResponse, error) {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return "", PlaceCardResponse{}, err
	}

	_, playerKey, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not find the card onwer %q, %s", cardID[:], err.Error()))
		return "", PlaceCardResponse{}, err
	}

	logger.LogDebug(fmt.Sprintf("[backend] owner of the card %s", playerKey))

	_, playerOneKey, err := GetPlayerOneFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could error getting player one table from , %s: %s", gameKey, err.Error()))
		return "", PlaceCardResponse{}, err
	}
	isPlayerOne := strings.Contains(playerOneKey, strings.ToLower(walletAddress[2:]))
	if isPlayerOne {
		logger.LogInfo("[backend] the player is player one")
	} else {
		logger.LogInfo("[backend] the player is player two")
	}

	p1Cards, p2Cards, err := GetPlacedCardsFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get placed cards %s, %s", gameKey, err.Error()))
		return "", PlaceCardResponse{}, err
	}

	if isPlayerOne {
		p1Cards++
	} else {
		p2Cards++
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return "", PlaceCardResponse{}, err
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
			Table: "PlacedCards",
			Key:   gameKey,
			Fields: []data.Field{
				{Key: "p1Cards", Data: data.UintField{Data: *big.NewInt(p1Cards)}},
				{Key: "p2Cards", Data: data.UintField{Data: *big.NewInt(p2Cards)}},
			},
		},
		{
			Table: "CurrentMana",
			Key:   gameKey,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - summonManaCost)}},
			},
		},
	}

	cardIDAsString := hexutil.Encode(cardID[:])
	if cardAbilityType, err := GetCardAbilityType(db, w, cardIDAsString); err != nil {
		// Store sidestep initial position
		if cardAbilityType == abilitySidestep {
			if pos, err := GetCardPosition(db, w, cardIDAsString); err != nil {
				events = append(events, data.MudEvent{
					Table: "SidestepInitialPosition",
					Key:   cardIDAsString,
					Fields: []data.Field{
						{Key: "placed", Data: data.BoolField{Data: true}},
						{Key: "x", Data: data.UintField{Data: *big.NewInt(pos.X)}},
						{Key: "y", Data: data.UintField{Data: *big.NewInt(pos.Y)}},
					},
				})
			}
		}
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: events,
	},
	)

	response := PlaceCardResponse{
		UUID:         msg.UUID,
		MsgType:      "placecardresponse",
		CardID:       msg.CardID,
		X:            msg.X,
		Y:            msg.Y,
		Player:       playerKey,
		LeftOverMana: currentMana - summonManaCost,
	}

	return gameKey, response, nil
}

func PlaceCardHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, PlaceCardResponse, error) {
	if !authenticated {
		return "", PlaceCardResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing place card request")

	var msg PlaceCard
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding place card message: %s", err))
		return "", PlaceCardResponse{}, nil
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to place card: %s", err))
		return "", PlaceCardResponse{}, nil
	}

	valid, err := validatePlaceCard(db, cardID, &msg, walletAddress)
	if err != nil || !valid {
		return "", PlaceCardResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "placecard", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to place card: %s", err))
		return "", PlaceCardResponse{}, nil
	}

	// TODO: maybe if this fails stop accepting transactions until a new block is created
	gameID, response, err := placeCardPrediction(db, cardID, &msg, txhash, walletAddress)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction place card: %s", err))
		return "", PlaceCardResponse{}, err
	}
	return gameID, response, nil
}
