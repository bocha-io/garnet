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

func attackPrediction(db *data.Database, cardID [32]byte, msg *Attack, txhash common.Hash) error {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return err
	}

	attackedCard, err := GetCardInPosition(db, w, gameKey, msg.X, msg.Y)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] there is no card to attack %s, (%d,%d): %s", gameKey, msg.X, msg.Y, err.Error()))
		return err
	}

	logger.LogInfo(fmt.Sprintf("[backend] attaking the card %s", attackedCard))
	attackDmg, err := GetCardAttack(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get the attack dmg for card %q, %s", cardID, err.Error()))
		return err
	}

	currentHp, err := GetCardCurrentHp(db, w, attackedCard)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get the current hp for card %s, %s", attackedCard, err.Error()))
		return err
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return err
	}

	events := []data.MudEvent{
		{
			Table: "ActionReady",
			Key:   hexutil.Encode(cardID[:]),
			Fields: []data.Field{
				{Key: "value", Data: data.BoolField{Data: false}},
			},
		},
		{
			Table: "CurrentMana",
			Key:   gameKey,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - attackManaCost)}},
			},
		},
	}

	if currentHp <= attackDmg {
		events = append(events, data.MudEvent{
			Table: "CurrentHp",
			Key:   attackedCard,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(0)}},
			},
		})
		events = append(events, data.MudEvent{
			Table: "Position",
			Key:   attackedCard,
			Fields: []data.Field{
				{Key: "placed", Data: data.BoolField{Data: true}},
				{Key: "gamekey", Data: gameField},
				{Key: "x", Data: data.UintField{Data: *big.NewInt(99)}},
				{Key: "y", Data: data.UintField{Data: *big.NewInt(99)}},
			},
		})
	} else {
		events = append(events, data.MudEvent{
			Table: "CurrentHp",
			Key:   attackedCard,
			Fields: []data.Field{
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentHp - attackDmg)}},
			},
		})
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: events,
	},
	)

	return nil
}

func AttackHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) {
	// TODO: Wallet address is used to validate the action
	_ = walletAddress
	if !authenticated {
		return
	}

	logger.LogDebug("[backend] processing attack request")

	var msg Attack
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding attack message: %s", err))
		return
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack a card: %s", err))
		return
	}

	txhash, err := txbuilder.SendTransaction(walletID, "attack", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack: %s", err))
	}
	_ = attackPrediction(db, cardID, &msg, txhash)
}
