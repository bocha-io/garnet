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

func movePrediction(db *data.Database, cardID [32]byte, msg *MoveCard, txhash common.Hash) error {
	w := db.GetWorld(WorldID)
	gameField, gameKey, err := GetGameFromCard(db, w, cardID)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] card not found in used in table %q, %s", cardID[:], err.Error()))
		return err
	}

	currentMana, err := GetCurrentManaFromGame(db, w, gameKey)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] could not get current mana %s, %s", gameKey, err.Error()))
		return err
	}

	db.AddTxSent(data.UnconfirmedTransaction{
		Txhash: txhash.Hex(),
		Events: []data.MudEvent{
			{
				Table: "Position",
				Key:   hexutil.Encode(cardID[:]),
				Fields: []data.Field{
					{Key: "placed", Data: data.BoolField{Data: true}},
					{Key: "gamekey", Data: gameField},
					{Key: "x", Data: data.UintField{Data: *big.NewInt(msg.X)}},
					{Key: "y", Data: data.UintField{Data: *big.NewInt(msg.Y)}},
				},
			},
			{
				Table: "CurrentMana",
				Key:   gameKey,
				Fields: []data.Field{
					{Key: "value", Data: data.UintField{Data: *big.NewInt(currentMana - moveManaCost)}},
				},
			},
		},
	},
	)

	return nil
}

func MoveHandler(authenticated bool, walletID int, walletAddress string, db *data.Database, p []byte) {
	// TODO: Wallet address is used to validate the action
	_ = walletAddress
	if !authenticated {
		return
	}

	logger.LogDebug("[backend] processing move card request")

	var msg MoveCard
	err := json.Unmarshal(p, &msg)
	if err != nil {
		logger.LogError(fmt.Sprintf("[backend] error decoding move card message: %s", err))
		return
	}

	cardID, err := dbconnector.StringToSlice(msg.CardID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to move card: %s", err))
		return
	}

	txhash, err := txbuilder.SendTransaction(walletID, "movecard", cardID, uint32(msg.X), uint32(msg.Y))
	if err != nil {
		// TODO: send response saying that the game could not be created
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to move card: %s", err))
	}
	_ = movePrediction(db, cardID, &msg, txhash)
}
