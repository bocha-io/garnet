package mudhandlers

import (
	"fmt"

	"github.com/bocha-io/garnet/x/indexer/data"
	"github.com/bocha-io/garnet/x/indexer/data/mudhelpers"
	"github.com/bocha-io/garnet/x/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"go.uber.org/zap"
)

func HandleGenericTableEvent(event *mudhelpers.StorecoreStoreSetRecord, db *data.Database) data.MudEvent {
	tableID := mudhelpers.PaddedTableId(event.TableId)
	logger.LogDebug(
		fmt.Sprintln(
			"handling generic table event",
			zap.String("world_address", event.WorldAddress()),
			zap.String("table_id", tableID),
		),
	)

	table := db.GetTable(event.WorldAddress(), tableID)

	// Decode the row record data
	fields := data.BytesToFields(event.Data, *table.Schema.Schema.Value, table.Schema.FieldNames)

	// Decode the row key
	aggregateKey := data.AggregateKey(event.Key)

	// Debug info
	a := ""
	for _, v := range *fields {
		a = fmt.Sprintf("%s. %s (%s)", a, v.String(), v.Type())
	}
	logger.LogDebug(fmt.Sprintf("[indexer] generic table event (%s) %s, key = %s, fields = %s", table.Metadata.TableID, table.Metadata.TableName, hexutil.Encode(aggregateKey), a))

	// Save it
	return db.AddRow(table, aggregateKey, fields)

	// TODO: do we need the info of each key or is it always going to match the complete expresion
	// decodedKeyData := mudhelpers.DecodeData(aggregateKey, *table.Schema.Schema.Key)
	// decodedKeyDataNew := data.BytesToFields(aggregateKey, *table.Schema.Schema.Key, table.Schema.KeyNames)
}
