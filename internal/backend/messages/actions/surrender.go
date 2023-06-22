package actions

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bocha-io/garnet/internal/backend/messages/dbconnector"
	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/bocha-io/garnet/internal/txbuilder"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func validateSurrender(db *data.Database, gameID [32]byte, walletAddress string) (bool, error) {
	if len(walletAddress) > 2 {
		walletAddress = walletAddress[2:]
	}
	w := db.GetWorld(WorldID)
	gameKey := hexutil.Encode(gameID[:])
	if !IsMatchCreated(db, w, gameKey) {
		return false, fmt.Errorf("match does not exist")
	}

	_, currentPlayer, err := GetCurrentPlayerFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the current player %s: %s", gameKey, err.Error()))
		return false, err
	}

	if !strings.Contains(currentPlayer, walletAddress) {
		logger.LogError(fmt.Sprintf("[backend] current player different from player (%s, %s)", currentPlayer, walletAddress))
		return false, fmt.Errorf("not player turn")
	}
	return true, nil
}

func SurrenderHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) (string, SurrenderResponse, error) {
	if !authenticated {
		return "", SurrenderResponse{}, fmt.Errorf("user not authenticated")
	}

	logger.LogDebug("[backend] processing surrender request")

	var msg Surrender
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding surrender message: %s", err))
		return "", SurrenderResponse{}, err
	}

	logger.LogDebug(fmt.Sprintf("[backend] creating endturn tx: %s", msg.MatchID))

	matchID, err := dbconnector.StringToSlice(msg.MatchID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to end turn: %s", err))
		return "", SurrenderResponse{}, err
	}

	valid, err := validateSurrender(db, matchID, walletAddress)
	if err != nil || !valid {
		logger.LogDebug(fmt.Sprintf("[backend] error invalid end turn: %s", err))
		return "", SurrenderResponse{}, err
	}

	_, err = txbuilder.SendTransaction(walletID, "surrender", matchID)
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to endturn: %s", err))
		return "", SurrenderResponse{}, err
	}

	key := strings.Replace(walletAddress, "0x", "0x000000000000000000000000", 1)
	w := db.GetWorld(WorldID)
	_, name, err := GetUserName(db, w, key)
	if err != nil {
		logger.LogError("[backend] match does not have a player one")
		return "", SurrenderResponse{}, err
	}
	nameString, err := hexutil.Decode(name)
	if err != nil {
		logger.LogError("[backend] could not decode players name")
		return "", SurrenderResponse{}, err
	}

	return msg.MatchID, SurrenderResponse{
		UUID:          msg.UUID,
		MsgType:       "surrenderresponse",
		Loser:         walletAddress,
		LoserUsername: string(nameString),
	}, nil
}
