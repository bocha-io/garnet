package indexer

import (
	"context"
	"math/big"
	"time"
	"fmt"

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
       startingHeight := 0
       endHeight := height

       if height > uint64(startingHeight)+500 {
	       endHeight = uint64(startingHeight) + 500
       }

       eth.ProcessBlocks(c, database, big.NewInt(int64(startingHeight)), big.NewInt(int64(endHeight)))

	// eth.ProcessBlocks(c, database, nil, big.NewInt(int64(height)))

	for !*quit {
		newHeight, err := c.BlockNumber(context.Background())
		if err != nil {
			logger.LogError("could not get the latest height")
			// TODO: retry instead of panic
			panic("")
		}

	      if newHeight != endHeight {
		       startingHeight = int(endHeight)
		       endHeight = uint64(newHeight)

		       if newHeight > uint64(startingHeight)+500 {
			       endHeight = uint64(startingHeight) + 500
		       }

		      logger.LogInfo(fmt.Sprintf("Heights: %d %d", startingHeight, endHeight))

		       eth.ProcessBlocks(c, database, big.NewInt(int64(startingHeight)), big.NewInt(int64(endHeight)))
		}

		database.LastHeight = newHeight

		time.Sleep(100 * time.Millisecond)
	}
}
