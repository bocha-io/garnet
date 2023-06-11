package dbconnector

import (
	"fmt"

	"github.com/bocha-io/garnet/internal/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func StringToSlice(stringID string) ([32]byte, error) {
	id, err := hexutil.Decode(stringID)
	if err != nil {
		logger.LogDebug(fmt.Sprintf("[backend] error creating transaction to place card: %s", err))
		return [32]byte{}, fmt.Errorf("error decoding the string %s", err.Error())
	}

	if len(id) != 32 {
		logger.LogDebug("[backend] error creating transaction to place card: invalid length")
		return [32]byte{}, fmt.Errorf("invalid length")
	}

	var idArray [32]byte
	copy(idArray[:], id)
	return idArray, nil
}
