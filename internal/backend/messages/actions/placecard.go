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

func placeCardPrediction(db *data.Database, cardID [32]byte, msg *PlaceCard, txhash common.Hash, walletAddress string) error {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return err
	}

	_, playerKey, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not find the card onwer %q, %s", cardID[:], err.Error()))
		return err
	}

	logger.LogDebug(fmt.Sprintf("[backend] owner of the card %s", playerKey))

	_, playerOneKey, err := GetPlayerOneFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could error getting player one table from , %s: %s", gameKey, err.Error()))
		return err
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
		return err
	}

	if isPlayerOne {
		p1Cards++
	} else {
		p2Cards++
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return err
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: []data.MudEvent{
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
		},
	},
	)

	return nil
}

func PlaceCardHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) {
	if !authenticated {
		return
	}

	logger.LogDebug("[backend] processing place card request")

	var msg PlaceCard
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding place card message: %s", err))
		return
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to place card: %s", err))
		return
	}

	// TODO: validate place card action before sending the transaction
	txhash, err := txbuilder.SendTransaction(walletID, "placecard", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to place card: %s", err))
	}

	// TODO: maybe if this fails stop accepting transactions until a new block is created
	_ = placeCardPrediction(db, cardID, &msg, txhash, walletAddress)
}
