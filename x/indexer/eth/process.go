package eth

import (
	"fmt"
	"math/big"

	"github.com/bocha-io/ethclient/x/ethclient"
	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/indexer/data/mudhelpers"
	"github.com/bocha-io/garnet/x/indexer/eth/mudhandlers"
	"github.com/bocha-io/logger"
)

type UnconfirmedTransaction struct {
	Txhash string
	Events *[]data.MudEvent
}

func ProcessBlocks(c *ethclient.EthClient, db *data.Database, initBlockHeight *big.Int, endBlockHeight *big.Int) {
	logs := c.FilterLogs(QueryForStoreLogs(initBlockHeight, endBlockHeight))
	logs = OrderLogs(logs)
	logger.LogInfo(fmt.Sprintf("[indexer] processing logs up to %d", endBlockHeight))

	processedTxns := map[string]*UnconfirmedTransaction{}

	for _, v := range logs {
		found := false
		if _, ok := processedTxns[v.TxHash.Hex()]; ok {
			found = true
		} else {
			for k, txsent := range db.UnconfirmedTransactions {
				if v.TxHash.Hex() == txsent.Txhash {
					logger.LogInfo(fmt.Sprintf("[indexer] procesing tx from mempool with hash %s", txsent))
					temp := db.UnconfirmedTransactions[k]
					db.UnconfirmedTransactions = append(db.UnconfirmedTransactions[:k], db.UnconfirmedTransactions[k+1:]...)
					processedTxns[temp.Txhash] = &UnconfirmedTransaction{Txhash: temp.Txhash, Events: &temp.Events}
					found = true
					break
				}
			}
		}

		var logMudEvent data.MudEvent

		if v.Topics[0].Hex() == mudhelpers.GetStoreAbiEventID("StoreSetRecord").Hex() {
			event, err := mudhandlers.ParseStoreSetRecord(v)
			if err != nil {
				logger.LogError(fmt.Sprintf("[indexer] error decoding message:%s", err))
				// TODO: what should we do here?
				break
			}
			switch mudhelpers.PaddedTableId(event.TableId) {
			case mudhelpers.SchemaTableId():
				logger.LogInfo("[indexer] processing and creating schema table")
				mudhandlers.HandleSchemaTableEvent(event, db)
			case mudhelpers.MetadataTableId():
				logger.LogInfo("[indexer] processing and updating a schema with metadata")
				mudhandlers.HandleMetadataTableEvent(event, db)
			default:
				logger.LogInfo("[indexer] processing a generic table event like adding a row")
				logMudEvent = mudhandlers.HandleGenericTableEvent(event, db)
			}
		}

		if v.Topics[0].Hex() == mudhelpers.GetStoreAbiEventID("StoreSetField").Hex() {
			event, err := mudhandlers.ParseStoreSetField(v)
			logger.LogInfo("[indexer] processing store set field message")
			if err != nil {
				logger.LogError(fmt.Sprintf("[indexer] error decoding message for store set field:%s\n", err))
			} else {
				logMudEvent = mudhandlers.HandleSetFieldEvent(event, db)
			}
		}
		if v.Topics[0].Hex() == mudhelpers.GetStoreAbiEventID("StoreDeleteRecord").Hex() {
			logger.LogInfo("[indexer] processing store delete record message")
			event, err := mudhandlers.ParseStoreDeleteRecord(v)
			if err != nil {
				logger.LogError(fmt.Sprintf("[indexer] error decoding message for store delete record:%s\n", err))
			} else {
				logMudEvent = mudhandlers.HandleDeleteRecordEvent(event, db)
			}
		}

		if found {
			fmt.Printf("validating table:%s, key:%s\n", logMudEvent.Table, logMudEvent.Key)
			for i, event := range *processedTxns[v.TxHash.Hex()].Events {
				if logMudEvent.Table == event.Table && logMudEvent.Key == event.Key {

					for j, field := range event.Fields {
						if logMudEvent.Fields[j].Data.String() != field.Data.String() {
							fmt.Printf("fieldA:%s.fieldB:%s. %s != %s, for table %s, id %s\n", logMudEvent.Fields[j].Key, field.Key, logMudEvent.Fields[j].Data.String(), field.Data.String(), logMudEvent.Table, logMudEvent.Key)
							panic("the prediction was wrong!")
						}
					}

					temp := *processedTxns[v.TxHash.Hex()].Events

					//nolint:gocritic
					if len(temp) == 1 {
						temp = []data.MudEvent{}
					} else if len(temp) == i {
						temp = temp[:i]
					} else {
						temp = append(temp[:i], temp[i+1:]...)
					}

					processedTxns[v.TxHash.Hex()].Events = &temp
					break
				}
			}
		}

	}

	for _, v := range processedTxns {
		if len(*v.Events) > 0 {
			fmt.Printf("events not processed %s %s %s", v.Txhash, (*v.Events)[0].Table, (*v.Events)[0].Key)
			panic("events were not proccessed")
		}
	}
}
