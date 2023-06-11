package actions

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/bocha-io/garnet/internal/backend/messages/dbconnector"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/bocha-io/garnet/internal/txbuilder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func validatePiercingShot(db *data.Database, cardID [32]byte, msg *PiercingShot, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)

	_, err := commonValidation(db, w, cardID, walletAddress, piercingShotManaCost)
	if err != nil {
		return false, err
	}

	cardAbilityType, err := GetCardAbilityType(db, w, hexutil.Encode(cardID[:]))
	if err != nil || cardAbilityType != abilityPiercingShot {
		return false, fmt.Errorf("card does not have the ability")
	}

	actionReady := IsCardReady(db, w, hexutil.Encode(cardID[:]))
	if !actionReady {
		logger.LogError(fmt.Sprintf("[backend] card already attacked: %s", hexutil.Encode(cardID[:])))
		return false, nil
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

	if !(((position.X == msg.X) && (position.Y == msg.Y-1 || position.Y == msg.Y+1)) ||
		((position.Y == msg.Y) && (position.X == msg.X-1 || position.X == msg.X+1))) {
		logger.LogError("[backend] attaking out of range")
		return false, err
	}

	return true, nil
}

func piercingShotPrediction(db *data.Database, cardID [32]byte, msg *PiercingShot, walletAddress string, txhash common.Hash) (string, SkillResponse, error) {
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

	position, err := GetCardPosition(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError("[backend] could not get the card position")
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
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - piercingShotManaCost)}},
			},
		},
	}

	// Attack
	pos := []data.Position{{X: position.X + 1, Y: position.Y}, {X: position.X + 2, Y: position.Y}, {X: position.X + 3, Y: position.Y}}
	if position.X == msg.X-1 { //nolint:gocritic
		pos = []data.Position{{X: position.X - 1, Y: position.Y}, {X: position.X - 2, Y: position.Y}, {X: position.X - 3, Y: position.Y}}
	} else if position.Y == msg.Y+1 {
		pos = []data.Position{{X: position.X, Y: position.Y + 1}, {X: position.X, Y: position.Y + 2}, {X: position.X, Y: position.Y + 3}}
	} else if position.Y == msg.Y-1 {
		pos = []data.Position{{X: position.X, Y: position.Y - 1}, {X: position.X, Y: position.Y - 2}, {X: position.X, Y: position.Y - 3}}
	}

	baseAlreadyAttacked := false
	affectedCards := []AffectedCard{}

	for _, v := range pos {
		attackedCard, err := GetCardInPosition(db, w, gameKey, v.X, v.Y)
		if err == nil {
			_, base, err := GetBaseFromCard(db, w, attackedCard)
			if err == nil {
				if baseAlreadyAttacked {
					continue
				}
				logger.LogDebug("[backend] the card is attacking the base")
				attackedCard = base
				baseAlreadyAttacked = true
			}

			_, attackedOwner, err := GetCardOwnerWithString(db, w, attackedCard)
			if err != nil {
				continue
			}

			if strings.Contains(attackedOwner, walletAddress) {
				// Friendly fire not allowed
				continue
			}

			if cover, err := GetCoverCard(db, w, gameKey, cardOwner, v.X, v.Y); err == nil {
				attackedCard = cover
			}

			logger.LogInfo(fmt.Sprintf("[backend] attaking the card %s", attackedCard))
			attackDmg := piercingShotAttackDmg

			currentHp, err := GetCardCurrentHp(db, w, attackedCard)
			if err != nil {
				logger.LogError(fmt.Sprintf("[backend] could not get the current hp for card %s, %s", attackedCard, err.Error()))
				return "", SkillResponse{}, err
			}

			if currentHp <= attackDmg {
				affectedCards = append(affectedCards, AffectedCard{CardIDAttacked: attackedCard, PreviousHp: currentHp, CurrentHp: 0})
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
				affectedCards = append(affectedCards, AffectedCard{CardIDAttacked: attackedCard, PreviousHp: currentHp, CurrentHp: currentHp - attackDmg})
				events = append(events, data.MudEvent{
					Table: "CurrentHp",
					Key:   attackedCard,
					Fields: []data.Field{
						{Key: "value", Data: data.UintField{Data: *big.NewInt(currentHp - attackDmg)}},
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

	response := SkillResponse{
		MovedCards:     []MovedCard{},
		AffectedCards:  affectedCards,
		UUID:           msg.UUID,
		MsgType:        "piercingshotresponse",
		CardIDAttacker: msg.CardID,
		Player:         cardOwner,
		LeftOverMana:   currentMana - attackManaCost,
	}

	return gameKey, response, nil
}

func PiercingShotHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, SkillResponse, error) {
	// TODO: Wallet address is used to validate the action
	if !authenticated {
		return "", SkillResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing attack request")

	var msg PiercingShot
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

	valid, err := validatePiercingShot(db, cardID, &msg, walletAddress)
	if err != nil || !valid {
		return "", SkillResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "piercingshot", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack: %s", err))
		return "", SkillResponse{}, nil
	}

	gameID, response, err := piercingShotPrediction(db, cardID, &msg, walletAddress, txhash)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction attack: %s", err))
		return "", SkillResponse{}, err
	}

	return gameID, response, nil
}
