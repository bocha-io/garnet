package indexer

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/indexer/eth"
	"github.com/bocha-io/logger"
)

// func ProcessMempool(database *data.Database, quit *bool) {
// 	logger.LogInfo("processing mempool...")
// 	c := eth.GetEthereumClient("http://localhost:8545/")
//     c.PendingTransactionCount()
// }

func Process(endpoint string, database *data.Database, quit *bool, startingHeight uint64, sleepDuration time.Duration) {
	logger.LogInfo("indexer is starting...")
	c := eth.GetEthereumClient(endpoint)
	ctx := context.Background()
	chainID, err := c.ChainID(ctx)
	if err != nil {
		logger.LogError("could not get the latest height")
		// TODO: retry instead of panic
		panic("")
	}
	database.ChainID = chainID.String()

	height, err := c.BlockNumber(context.Background())
	if err != nil {
		logger.LogError("could not get the latest height")
		// TODO: retry instead of panic
		panic("")
	}

	endHeight := height

	if height > startingHeight+500 {
		endHeight = startingHeight + 500
	}

	eth.ProcessBlocks(c, database, big.NewInt(int64(startingHeight)), big.NewInt(int64(endHeight)))

	for !*quit {
		newHeight, err := c.BlockNumber(context.Background())
		if err != nil {
			logger.LogError("could not get the latest height")
			// TODO: retry instead of panic
			panic("")
		}

		if newHeight != endHeight {
			startingHeight = endHeight
			endHeight = newHeight

			if newHeight > startingHeight+500 {
				endHeight = startingHeight + 500
			}

			logger.LogInfo(fmt.Sprintf("Heights: %d %d", startingHeight, endHeight))

			eth.ProcessBlocks(c, database, big.NewInt(int64(startingHeight)), big.NewInt(int64(endHeight)))
		}

		database.LastHeight = newHeight

		time.Sleep(sleepDuration)
	}
}
