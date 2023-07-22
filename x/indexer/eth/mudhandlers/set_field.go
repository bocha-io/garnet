package mudhandlers

import (
	"fmt"

	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/indexer/data/mudhelpers"
	"github.com/bocha-io/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

func HandleSetFieldEvent(event *mudhelpers.StorecoreStoreSetField, db *data.Database) data.MudEvent {
	tableID := mudhelpers.PaddedTableId(event.TableId)
	logger.LogDebug(
		fmt.Sprintln(
			"handling set field (StoreSetFieldEvent) event",
			zap.String("table_id", tableID),
		),
	)

	table := db.GetTable(event.WorldAddress(), tableID)

	// Handle the following scenarios:
	// 1. The setField event is modifying a row that doesn't yet exist (i.e. key doesn't match anything),
	//    in which case we insert a new row with default values for each column.
	//
	// 2. The setField event is modifying a row that already exists, in which case we update the
	//    row by constructing a partial row with the new value for the field that was modified.

	key := data.AggregateKey(event.Key)

	mudevent := db.SetField(table, key, event)

	a := ""
	for _, v := range mudevent.Fields {
		a = fmt.Sprintf("%s. %s (%s)", a, v.String(), v.Type())
	}
	logger.LogDebug(fmt.Sprintf("[indexer] generic table event (%s) %s, key = %s, fields = %s", table.Metadata.TableID, table.Metadata.TableName, hexutil.Encode(key), a))
	return mudevent
}
