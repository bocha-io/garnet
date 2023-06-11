package actions

import (
	"encoding/json"
	"fmt"
	"math"
	"math/big"

	"github.com/bocha-io/garnet/internal/backend/messages/dbconnector"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/bocha-io/garnet/internal/txbuilder"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func validateCover(db *data.Database, cardID [32]byte, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)

	_, err := commonValidation(db, w, cardID, walletAddress, coverManaCost)
	if err != nil {
		return false, err
	}

	cardAbilityType, err := GetCardAbilityType(db, w, hexutil.Encode(cardID[:]))
	if err != nil || cardAbilityType != abilityCover {
		return false, fmt.Errorf("card does not have the ability")
	}

	actionReady := IsCardReady(db, w, hexutil.Encode(cardID[:]))
	if !actionReady {
		logger.LogError(fmt.Sprintf("[backend] card already attacked: %s", hexutil.Encode(cardID[:])))
		return false, nil
	}

	return true, nil
}

func coverPrediction(db *data.Database, cardID [32]byte, msg *Cover, txhash common.Hash) (string, SkillResponse, error) {
	w := db.GetWorld(WorldID)
	_, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return "", SkillResponse{}, err
	}

	owner, cardOwner, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the card owner %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", SkillResponse{}, err
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return "", SkillResponse{}, err
	}

	cover, err := GetCoverPosition(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] there is no cover  %s: %s", gameKey, err.Error()))
		return "", SkillResponse{}, err
	}

	newFields := cover.Raw
	if cardOwner == cover.Player { //nolint:gocritic
		newFields[0] = data.Field{Key: "coverOneCard", Data: data.NewBytesField(cardID[:])}
		newFields[1] = data.Field{Key: "coverOnePlayer", Data: owner.Data}
	} else if cardOwner == cover.Player2 {
		newFields[2] = data.Field{Key: "coverTowCard", Data: data.NewBytesField(cardID[:])}
		newFields[3] = data.Field{Key: "coverTowPlayer", Data: owner.Data}
	} else if cover.Player == emptyString {
		newFields[0] = data.Field{Key: "coverOneCard", Data: data.NewBytesField(cardID[:])}
		newFields[1] = data.Field{Key: "coverOnePlayer", Data: owner.Data}
	} else if cover.Player2 == emptyString {
		newFields[2] = data.Field{Key: "coverTowCard", Data: data.NewBytesField(cardID[:])}
		newFields[3] = data.Field{Key: "coverTowPlayer", Data: owner.Data}
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
				{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - coverManaCost)}},
			},
		},
		{
			Table:  "CoverPosition",
			Key:    gameKey,
			Fields: newFields,
		},
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: events,
	},
	)

	response := SkillResponse{
		UUID:           msg.UUID,
		MsgType:        "coverresponse",
		CardIDAttacker: msg.CardID,
		MovedCards:     []MovedCard{},
		AffectedCards:  []AffectedCard{},
		Player:         cardOwner,
		LeftOverMana:   currentMana - coverManaCost,
	}

	return gameKey, response, nil
}

func CoverHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, SkillResponse, error) {
	// TODO: Wallet address is used to validate the action
	if !authenticated {
		return "", SkillResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing attack request")

	var msg Cover
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

	valid, err := validateCover(db, cardID, walletAddress)
	if err != nil || !valid {
		return "", SkillResponse{}, nil
	}

	txhash, err := txbuilder.SendTransaction(walletID, "cover", cardID)
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to attack: %s", err))
		return "", SkillResponse{}, nil
	}

	gameID, response, err := coverPrediction(db, cardID, &msg, txhash)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error prediction attack: %s", err))
		return "", SkillResponse{}, err
	}

	return gameID, response, nil
}

func GetCoverCard(db *data.Database, w *data.World, gameKey string, player string, x int64, y int64) (string, error) {
	cover, err := GetCoverPosition(db, w, gameKey)
	if err != nil {
		return "", fmt.Errorf("could not get cover pos")
	}

	if cover.Player != emptyString && cover.Player != player {
		pos, err := GetCardPosition(db, w, cover.Card)
		if err != nil {
			return "", err
		}
		deltaX := math.Abs(float64(pos.X - x))
		deltaY := math.Abs(float64(pos.Y - y))
		if deltaX+deltaY > float64(coverRange) {
			return "", fmt.Errorf("out of range")
		}
		return cover.Card, nil
	} else if cover.Player2 != emptyString && cover.Player2 != player {
		pos, err := GetCardPosition(db, w, cover.Card2)
		if err != nil {
			return "", err
		}
		deltaX := math.Abs(float64(pos.X - x))
		deltaY := math.Abs(float64(pos.Y - y))
		if deltaX+deltaY > float64(coverRange) {
			return "", fmt.Errorf("out of range")
		}
		return cover.Card, nil
	}
	return "", fmt.Errorf("the cover is not set")
}
