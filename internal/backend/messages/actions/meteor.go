package actions

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/bocha-io/garnet/internal/backend/messages/dbconnector"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/bocha-io/garnet/internal/txbuilder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func validateMeteor(db *data.Database, cardID [32]byte, msg *Meteor, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)

	gameKey, err := commonValidation(db, w, cardID, walletAddress, meteorManaCost)
	if err != nil {
		return false, err
	}

	cardAbilityType, err := GetCardAbilityType(db, w, hexutil.Encode(cardID[:]))
	if err != nil || cardAbilityType != abilityMeteor {
		return false, fmt.Errorf("card does not have the ability")
	}

	actionReady := IsCardReady(db, w, hexutil.Encode(cardID[:]))
	if !actionReady {
		logger.LogError(fmt.Sprintf("[backend] card already attacked: %s", hexutil.Encode(cardID[:])))
		return false, nil
	}

	attackedKey, err := GetCardInPosition(db, w, gameKey, msg.X, msg.Y)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] there is no unit at position (%d, %d) game %s", msg.X, msg.Y, gameKey))
		return false, err
	}

	_, base, err := GetBaseFromCard(db, w, attackedKey)
	if err == nil {
		logger.LogDebug("[backend] the card is attacking the base")
		attackedKey = base
	}

	_, attackedOwner, err := GetCardOwnerWithString(db, w, attackedKey)
	if err != nil {
		logger.LogError("[backend] could not find the onwer of the attacked card")
		return false, err
	}

	if strings.Contains(attackedOwner, walletAddress) {
		logger.LogError("[backend] trying to attack its own card")
		return false, err
	}

	if msg.X < 0 || msg.X > 9 || msg.Y < 0 || msg.Y > 10 {
		logger.LogError("[backend] invalid position, is out of the map")
		return false, err
	}

	// Range
	position, err := GetCardPosition(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError("[backend] could not get the card position")
		return false, err
	}

	deltaX := math.Abs(float64(position.X - msg.X))
	deltaY := math.Abs(float64(position.Y - msg.Y))
	if deltaX+deltaY > float64(meteorRange) {
		logger.LogError("[backend] out of range")
		return false, nil
	}

	return true, nil
}

func meteorPrediction(db *data.Database, cardID [32]byte, msg *Meteor, txhash common.Hash) (string, SkillResponse, error) {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return "", SkillResponse{}, err
	}

	_, cardOwner, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the card owner %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", SkillResponse{}, err
	}

	attackedCard, err := GetCardInPosition(db, w, gameKey, msg.X, msg.Y)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] there is no card to attack %s, (%d,%d): %s", gameKey, msg.X, msg.Y, err.Error()))
		return "", SkillResponse{}, err
	}

	logger.LogInfo(fmt.Sprintf("[backend] attaking the card %s", attackedCard))
	attackDmg := int64(4)

	_, base, err := GetBaseFromCard(db, w, attackedCard)
	if err == nil {
		logger.LogDebug("[backend] the card is attacking the base")
		attackedCard = base
	}

	if cover, err := GetCoverCard(db, w, gameKey, cardOwner, msg.X, msg.Y); err == nil {
		attackedCard = cover
	}

	currentHp, err := GetCardCurrentHp(db, w, attackedCard)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get the current hp for card %s, %s", attackedCard, err.Error()))
		return "", SkillResponse{}, err
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return "", SkillResponse{}, err
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
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - meteorManaCost)}},
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

	response := SkillResponse{
		UUID:           msg.UUID,
		MsgType:        "meteorresponse",
		CardIDAttacker: msg.CardID,
		MovedCards:     []MovedCard{},
		AffectedCards: []AffectedCard{{
			CardIDAttacked: attackedCard,
			PreviousHp:     currentHp,
			CurrentHp:      currentHp - attackDmg,
		}},
		Player:       cardOwner,
		LeftOverMana: currentMana - meteorManaCost,
	}

	return gameKey, response, nil
}

func MeteorHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, SkillResponse, error) {
	// TODO: Wallet address is used to validate the action
	if !authenticated {
		return "", SkillResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing attack request")

	var msg Meteor
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding attack message: %s", err))
		return "", SkillResponse{}, nil
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack a card: %s", err))
		return "", SkillResponse{}, nil
	}

	valid, err := validateMeteor(db, cardID, &msg, walletAddress)
	if err != nil || !valid {
		return "", SkillResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "meteor", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack: %s", err))
		return "", SkillResponse{}, nil
	}

	gameID, response, err := meteorPrediction(db, cardID, &msg, txhash)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction attack: %s", err))
		return "", SkillResponse{}, err
	}

	return gameID, response, nil
}