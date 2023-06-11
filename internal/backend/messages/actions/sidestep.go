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

func validateSidestep(db *data.Database, cardID [32]byte, msg *Sidestep, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)

	gameKey, err := commonValidation(db, w, cardID, walletAddress, sidestepManaCost)
	if err != nil {
		return false, err
	}

	cardAbilityType, err := GetCardAbilityType(db, w, hexutil.Encode(cardID[:]))
	if err != nil || cardAbilityType != abilitySidestep {
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

	if !(((position.X == msg.X) && (position.Y == msg.Y-1 || position.Y == msg.Y+1)) ||
		((position.Y == msg.Y) && (position.X == msg.X-1 || position.X == msg.X+1))) {
		logger.LogError("[backend] attaking out of range")
		return false, err
	}

	return true, nil
}

func sidestepPrediction(db *data.Database, cardID [32]byte, msg *Sidestep, txhash common.Hash) (string, SkillResponse, error) {
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
	attackDmg := sidestepAttack

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
	initialPos, err := GetCardSidestepInitialPosition(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get initial pos %s, %s", hexutil.Encode(cardID[:]), err.Error()))
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
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - sidestepManaCost)}},
			},
		},
		{
			Table: "Position",
			Key:   hexutil.Encode(cardID[:]),
			Fields: []data.Field{
				{Key: "placed", Data: data.BoolField{Data: true}},
				{Key: "gamekey", Data: gameField},
				{Key: "x", Data: data.UintField{Data: *big.NewInt(initialPos.X)}},
				{Key: "y", Data: data.UintField{Data: *big.NewInt(initialPos.Y)}},
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
		MovedCards:     []MovedCard{{CardID: msg.CardID, X: initialPos.X, Y: initialPos.Y}},
		UUID:           msg.UUID,
		MsgType:        "sidestepresponse",
		CardIDAttacker: msg.CardID,
		AffectedCards: []AffectedCard{
			{
				CardIDAttacked: attackedCard,
				PreviousHp:     currentHp,
				CurrentHp:      currentHp - attackDmg,
			},
		},
		Player:       cardOwner,
		LeftOverMana: currentMana - drainSwordManaCost,
	}

	return gameKey, response, nil
}

func SidestepHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, SkillResponse, error) {
	// TODO: Wallet address is used to validate the action
	if !authenticated {
		return "", SkillResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing attack request")

	var msg Sidestep
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

	valid, err := validateSidestep(db, cardID, &msg, walletAddress)
	if err != nil || !valid {
		return "", SkillResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "sidestep", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack: %s", err))
		return "", SkillResponse{}, nil
	}

	gameID, response, err := sidestepPrediction(db, cardID, &msg, txhash)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction attack: %s", err))
		return "", SkillResponse{}, err
	}

	return gameID, response, nil
}
