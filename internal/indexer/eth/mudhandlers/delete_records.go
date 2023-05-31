package mudhandlers

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/hanchon/garnet/internal/indexer/data"
	"github.com/hanchon/garnet/internal/indexer/data/mudhelpers"
	"github.com/hanchon/garnet/internal/logger"
	"go.uber.org/zap"
)

func HandleDeleteRecordEvent(event *mudhelpers.StorecoreStoreDeleteRecord, db *data.Database) data.MudEvent {
	tableID := mudhelpers.PaddedTableId(event.TableId)
	logger.LogDebug(
		fmt.Sprintln(
			"handling delete record event",
			zap.String("table_id", tableID),
		),
	)

	table := db.GetTable(event.WorldAddress(), tableID)

	aggregateKey := data.AggregateKey(event.Key)

	logger.LogDebug(fmt.Sprintf("[indexer] deleting element from table (%s) %s, key = %s", table.Metadata.TableID, table.Metadata.TableName, hexutil.Encode(aggregateKey)))

	return db.DeleteRow(table, aggregateKey)
}
