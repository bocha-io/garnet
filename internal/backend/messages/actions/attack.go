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

func attackPrediction(db *data.Database, cardID [32]byte, msg *Attack, txhash common.Hash) (string, AttackResponse, error) {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return "", AttackResponse{}, err
	}

	_, cardOwner, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the card owner %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", AttackResponse{}, err
	}

	attackedCard, err := GetCardInPosition(db, w, gameKey, msg.X, msg.Y)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] there is no card to attack %s, (%d,%d): %s", gameKey, msg.X, msg.Y, err.Error()))
		return "", AttackResponse{}, err
	}

	logger.LogInfo(fmt.Sprintf("[backend] attaking the card %s", attackedCard))
	attackDmg, err := GetCardAttack(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get the attack dmg for card %q, %s", cardID, err.Error()))
		return "", AttackResponse{}, err
	}

	currentHp, err := GetCardCurrentHp(db, w, attackedCard)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get the current hp for card %s, %s", attackedCard, err.Error()))
		return "", AttackResponse{}, err
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return "", AttackResponse{}, err
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

	response := AttackResponse{
		UUID:           msg.UUID,
		MsgType:        "attackresponse",
		CardIDAttacker: msg.CardID,
		CardIDAttacked: attackedCard,
		X:              msg.X,
		Y:              msg.Y,
		PreviousHp:     currentHp,
		CurrentHp:      currentHp - attackDmg,
		Player:         cardOwner,
		LeftOverMana:   currentMana - attackManaCost,
	}

	return gameKey, response, nil
}

func AttackHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, AttackResponse, error) {
	// TODO: Wallet address is used to validate the action
	_ = walletAddress
	if !authenticated {
		return "", AttackResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing attack request")

	var msg Attack
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding attack message: %s", err))
		return "", AttackResponse{}, nil
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack a card: %s", err))
		return "", AttackResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "attack", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack: %s", err))
		return "", AttackResponse{}, nil
	}
	gameID, response, err := attackPrediction(db, cardID, &msg, txhash)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction attack: %s", err))
		return "", AttackResponse{}, err
	}
	return gameID, response, nil
}
