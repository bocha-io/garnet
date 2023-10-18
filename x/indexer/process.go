package indexer

import (
	"fmt"
	"math/big"
	"time"

	"github.com/bocha-io/ethclient/x/ethclient"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/indexer/eth"
	"github.com/bocha-io/logger"
)

// func ProcessMempool(database *data.Database, quit *bool) {
// 	logger.LogInfo("processing mempool...")
// 	c := eth.GetEthereumClient("http://localhost:8545/")
//     c.PendingTransactionCount()
// }

func Process(client *ethclient.EthClient, database *data.Database, quit *bool, startingHeight uint64, sleepDuration time.Duration) {
	logger.LogInfo("indexer is starting...")
	database.ChainID = client.ChainID().String()

	height := client.BlockNumber()

	endHeight := height
	amountOfBlocks := uint64(500)

	if height > startingHeight+amountOfBlocks {
		endHeight = startingHeight + amountOfBlocks
	}

	eth.ProcessBlocks(client, database, big.NewInt(int64(startingHeight)), big.NewInt(int64(endHeight)))

	for !*quit {
		newHeight := client.BlockNumber()

		if newHeight != endHeight {
			startingHeight = endHeight
			endHeight = newHeight

			if newHeight > startingHeight+amountOfBlocks {
				endHeight = startingHeight + amountOfBlocks
			}

			logger.LogInfo(fmt.Sprintf("Heights: %d %d", startingHeight, endHeight))

			eth.ProcessBlocks(client, database, big.NewInt(int64(startingHeight)), big.NewInt(int64(endHeight)))
		}

		database.LastHeight = newHeight

		time.Sleep(sleepDuration)
	}
}
