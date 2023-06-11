package actions

import (
	"fmt"

	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
)

const baseType = 6

func GetBoard(db *data.Database, w *data.World, matchID string) *data.MatchData {
	logger.LogInfo("[backend] getting board status...")
	ret := data.MatchData{MatchID: matchID}

	_, playerOne, err := GetPlayerOneFromGame(db, w, matchID)
	if err != nil {
		logger.LogError("[backend] match does not have a player one")
		return nil
	}
	ret.PlayerOne = playerOne

	_, playerTwo, err := GetPlayerTwoFromGame(db, w, matchID)
	if err != nil {
		logger.LogError("[backend] match does not have a player one")
		return nil
	}
	ret.PlayerTwo = playerTwo

	// CurrentTurn
	currentTurn, err := GetCurrentTurnFromGame(db, w, matchID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current turn %s", err.Error()))
		return nil
	}
	ret.CurrentTurn = currentTurn

	// CurrentPlayer
	_, currentPlayer, err := GetCurrentPlayerFromGame(db, w, matchID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current player: %s", err.Error()))
		return nil
	}
	ret.CurrentPlayer = currentPlayer

	// CurrentMana
	currentMana, err := GetCurrentManaFromGame(db, w, matchID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s", err.Error()))
		return nil
	}
	ret.CurrentMana = currentMana

	// Cards
	matchCards := GetCardsFromMatch(db, w, matchID)
	cards := []data.Card{}

	for _, card := range matchCards {
		c := data.Card{ID: card}
		// IsBase
		if IsCardBase(db, w, card) {
			continue
		}

		// Owner
		_, cardOwner, err := GetCardOwnerUsingString(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] ownedby  error %s", err.Error()))
			return nil
		}
		c.Owner = cardOwner

		// MaxHp
		maxHp, err := GetCardMaxHp(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] max hp error %s", err.Error()))
			return nil
		}
		c.MaxHp = maxHp

		// CurrentHp
		currentHp, err := GetCardCurrentHp(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] current hp error %s", err.Error()))
			return nil
		}
		c.CurrentHp = currentHp

		// UnitType
		unitType, err := GetCardUnitType(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] unit type error %s", err.Error()))
			return nil
		}
		c.Type = unitType

		if unitType == baseType {
			c.Placed = false
			c.Position = data.Position{X: -2, Y: -2}
			cards = append(cards, c)
			continue
		}

		// AttackDamage
		attackDmg, err := GetCardAttackUsingString(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] attack damage error %s", err.Error()))
			return nil
		}
		c.AttackDamage = attackDmg

		// MovementSpeed
		movementSpeed, err := GetCardMovementSpeed(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] movement speed error %s", err.Error()))
			return nil
		}
		c.MovementSpeed = movementSpeed

		// ActionReady
		c.ActionReady = IsCardReady(db, w, card)

		// Position
		p, err := GetCardPosition(db, w, card)
		if err != nil {
			logger.LogInfo(fmt.Sprintf("[backend] get position error %s", err.Error()))
			return nil
		}

		if p.X == -2 && p.Y == -2 {
			c.Placed = false
		} else {
			c.Placed = true
		}

		c.Position = p

		cards = append(cards, c)
	}
	ret.Cards = cards

	return &ret
}

func GetBoardStatus(db *data.Database, worldID string, userWallet string) *data.MatchData {
	w, ok := db.Worlds[worldID]
	if !ok {
		logger.LogError("[backend] world not found")
		return nil
	}
	matchID := GetUserMatchID(db, w, userWallet)
	if matchID == "" {
		return nil
	}

	data := GetBoard(db, w, matchID)

	return data
}
