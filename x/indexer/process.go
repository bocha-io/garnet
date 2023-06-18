package indexer

import (
	"context"
	"math/big"
	"time"

	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/indexer/eth"
	"github.com/bocha-io/garnet/x/logger"
)

// func ProcessMempool(database *data.Database, quit *bool) {
// 	logger.LogInfo("processing mempool...")
// 	c := eth.GetEthereumClient("http://localhost:8545/")
//     c.PendingTransactionCount()
// }

func Process(endpoint string, database *data.Database, quit *bool) {
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

	eth.ProcessBlocks(c, database, nil, big.NewInt(int64(height)))

	for !*quit {
		newHeight, err := c.BlockNumber(context.Background())
		if err != nil {
			logger.LogError("could not get the latest height")
			// TODO: retry instead of panic
			panic("")
		}

		if newHeight != height {
			eth.ProcessBlocks(c, database, big.NewInt(int64(height)), big.NewInt(int64(newHeight)))
			height = newHeight
		}

		database.LastHeight = newHeight

		time.Sleep(1 * time.Second)
	}
}
