package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bocha-io/garnet/internal/backend/messages/dbconnector"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var emptyString = "0x0000000000000000000000000000000000000000000000000000000000000000"

func EmptyBytes() []byte {
	emptyBytes, _ := hexutil.Decode(emptyString)
	return emptyBytes
}

// Actual functions
func GetGameFromCard(db *data.Database, w *data.World, cardID [32]byte) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingBytes(db, w, cardID, "UsedIn")
}

func GetCardOwner(db *data.Database, w *data.World, cardID [32]byte) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingBytes(db, w, cardID, "OwnedBy")
}

func GetCardOwnerWithString(db *data.Database, w *data.World, cardID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, cardID, "OwnedBy")
}

func GetCardOwnerUsingString(db *data.Database, w *data.World, cardID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, cardID, "OwnedBy")
}

func GetPlayerOneFromGame(db *data.Database, w *data.World, gameID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, gameID, "PlayerOne")
}

func GetPlayerTwoFromGame(db *data.Database, w *data.World, gameID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, gameID, "PlayerTwo")
}

func GetUserName(db *data.Database, w *data.World, userID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, userID, "User")
}

// GetPlacedCardsFromGame returns p1cards, p2cards, error
func GetPlacedCardsFromGame(db *data.Database, w *data.World, gameID string) (int64, int64, error) {
	row, err := dbconnector.GetRowFieldsUsingString(db, w, gameID, "PlacedCards")
	if err != nil {
		return 0, 0, err
	}

	if len(row) != 2 {
		return 0, 0, fmt.Errorf("row does not have 2 parameters")
	}

	p1Cards, err := strconv.ParseInt(row[0].Data.String(), 10, 32)
	if err != nil {
		return 0, 0, err
	}

	p2cards, err := strconv.ParseInt(row[1].Data.String(), 10, 32)
	if err != nil {
		return 0, 0, err
	}
	return p1Cards, p2cards, err
}

func GetCurrentManaFromGame(db *data.Database, w *data.World, gameID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, gameID, "CurrentMana")
}

func GetCurrentTurnFromGame(db *data.Database, w *data.World, gameID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, gameID, "CurrentTurn")
}

func GetCurrentPlayerFromGame(db *data.Database, w *data.World, gameID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, gameID, "CurrentPlayer")
}

func IsMatchCreated(db *data.Database, w *data.World, gameID string) bool {
	_, value, err := dbconnector.GetRowFromIDUsingString(db, w, gameID, "Match")
	if err != nil || value != "true" {
		return false
	}
	return true
}

func GetCardInPosition(db *data.Database, w *data.World, gameID string, x int64, y int64) (string, error) {
	table := w.GetTableByName("Position")
	rows := db.GetRows(table)

	for k, v := range rows {
		if len(v) != 4 {
			break
		}
		if strings.Contains(v[1].Data.String(), gameID) {
			// Card in game
			if v[2].Data.String() == fmt.Sprint(x) && v[3].Data.String() == fmt.Sprint(y) {
				return k, nil
			}
		}
	}
	return "", fmt.Errorf("value not found")
}

func GetCardAttack(db *data.Database, w *data.World, cardID [32]byte) (int64, error) {
	return dbconnector.GetInt64UsingBytes(db, w, cardID, "AttackDamage")
}

func GetCardAttackUsingString(db *data.Database, w *data.World, cardID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, cardID, "AttackDamage")
}

func GetCardCurrentHp(db *data.Database, w *data.World, cardID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, cardID, "CurrentHp")
}

func GetCardMaxHp(db *data.Database, w *data.World, cardID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, cardID, "MaxHp")
}

func GetCardUnitType(db *data.Database, w *data.World, cardID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, cardID, "UnitType")
}

func GetCardAbilityType(db *data.Database, w *data.World, cardID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, cardID, "AbilityType")
}

func GetCardMovementSpeed(db *data.Database, w *data.World, cardID string) (int64, error) {
	return dbconnector.GetInt64UsingString(db, w, cardID, "MovementSpeed")
}

func GetBaseFromCard(db *data.Database, w *data.World, cardID string) (data.Field, string, error) {
	return dbconnector.GetRowFromIDUsingString(db, w, cardID, "IsBase")
}

func GetCardsFromMatch(db *data.Database, w *data.World, rowID string) []string {
	table := w.GetTableByName("UsedIn")
	rows := db.GetRows(table)

	IDs := []string{}
	for k, v := range rows {
		if strings.Contains(v[0].Data.String(), rowID) {
			IDs = append(IDs, k)
		}
	}
	return IDs
}

// TODO: this does not work!, the field is not boolean
func IsCardBase(db *data.Database, w *data.World, cardID string) bool {
	return dbconnector.GetBoolFromTable(db, w, cardID, "IsBase")
}

func IsCardReady(db *data.Database, w *data.World, cardID string) bool {
	return dbconnector.GetBoolFromTable(db, w, cardID, "ActionReady")
}

func GetCardPosition(db *data.Database, w *data.World, cardID string) (data.Position, error) {
	value, err := dbconnector.GetRowFieldsUsingString(db, w, cardID, "Position")
	if err != nil || len(value) != 4 {
		return data.Position{X: -2, Y: -2}, nil
	}

	x, err := strconv.ParseInt(value[2].Data.String(), 10, 32)
	if err != nil {
		errorMsg := fmt.Sprintf("[backend] could not parse X from %s value %s", "Position", value[2].Data.String())
		return data.Position{X: -2, Y: -2}, fmt.Errorf(errorMsg)
	}

	y, err := strconv.ParseInt(value[3].Data.String(), 10, 32)
	if err != nil {
		errorMsg := (fmt.Sprintf("[backend] could not parse X from %s value %s", "Position", value[3].Data.String()))
		return data.Position{X: -2, Y: -2}, fmt.Errorf(errorMsg)
	}

	if x == 99 || y == 99 {
		return data.Position{X: -2, Y: -2}, fmt.Errorf("the card was killed")
	}

	return data.Position{X: x, Y: y}, nil
}

func GetCardSidestepInitialPosition(db *data.Database, w *data.World, cardID string) (data.Position, error) {
	value, err := dbconnector.GetRowFieldsUsingString(db, w, cardID, "SidestepInitialPosition")
	if err != nil || len(value) != 3 {
		return data.Position{X: -2, Y: -2}, nil
	}

	x, err := strconv.ParseInt(value[1].Data.String(), 10, 32)
	if err != nil {
		errorMsg := fmt.Sprintf("[backend] could not parse X from %s value %s", "SidestepInitialPosition", value[2].Data.String())
		return data.Position{X: -2, Y: -2}, fmt.Errorf(errorMsg)
	}

	y, err := strconv.ParseInt(value[2].Data.String(), 10, 32)
	if err != nil {
		errorMsg := (fmt.Sprintf("[backend] could not parse X from %s value %s", "SidestepInitialPosition", value[3].Data.String()))
		return data.Position{X: -2, Y: -2}, fmt.Errorf(errorMsg)
	}

	return data.Position{X: x, Y: y}, nil
}

func GetCoverPosition(db *data.Database, w *data.World, gameID string) (data.CoverPosition, error) {
	value, err := dbconnector.GetRowFieldsUsingString(db, w, gameID, "CoverPosition")
	if err != nil || len(value) != 4 {
		return data.CoverPosition{}, fmt.Errorf("value not found")
	}

	return data.CoverPosition{
		Card:    strings.ReplaceAll(value[0].Data.String(), "\"", ""),
		Player:  strings.ReplaceAll(value[1].Data.String(), "\"", ""),
		Card2:   strings.ReplaceAll(value[2].Data.String(), "\"", ""),
		Player2: strings.ReplaceAll(value[3].Data.String(), "\"", ""),
		Raw:     value,
	}, nil
}

func GetUserMatchID(db *data.Database, w *data.World, userWallet string) string {
	// Check if the user is the player one
	playerOneRows := dbconnector.GetRows(db, w, "PlayerOne")
	for k, v := range playerOneRows {
		if strings.Contains(strings.ToLower(v[0].Data.String()), userWallet[2:]) {
			return k
		}
	}

	// Check if the user is the player two
	playerTwoRows := dbconnector.GetRows(db, w, "PlayerTwo")
	for k, v := range playerTwoRows {
		if strings.Contains(strings.ToLower(v[0].Data.String()), userWallet[2:]) {
			return k
		}
	}
	return ""
}
