package actions

import (
	"fmt"
	"strings"

	"github.com/bocha-io/garnet/internal/indexer/data"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Returns gameKey and error, if there is error, the validation failed
func commonValidation(db *data.Database, w *data.World, cardID [32]byte, walletAddress string, actionMana int64) (string, error) {
	_, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return "", err
	}

	_, cardOwner, err := GetCardOwner(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the card owner %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", err
	}

	if !strings.Contains(cardOwner, walletAddress) {
		logger.LogError(fmt.Sprintf("[backend] card owner different from player (%s, %s)", cardOwner, walletAddress))
		return "", fmt.Errorf("not player card")
	}

	_, currentPlayer, err := GetCurrentPlayerFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the current player %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", err
	}

	if !strings.Contains(currentPlayer, walletAddress) {
		logger.LogError(fmt.Sprintf("[backend] current player different from player (%s, %s)", currentPlayer, walletAddress))
		return "", fmt.Errorf("not player turn")
	}

	unitType, err := GetCardUnitType(db, w, hexutil.Encode(cardID[:]))
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the unity type %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", err
	}

	if unitType == baseType {
		logger.LogError("[backend] base can not attack")
		return "", fmt.Errorf("unit is a base")
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] failed to get the current mana %s: %s", hexutil.Encode(cardID[:]), err.Error()))
		return "", err
	}

	if currentMana < actionMana {
		return "", fmt.Errorf("not enough mana")
	}

	return gameKey, nil
}
