package data

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type MudEvent struct {
	Table  string
	Key    string
	Fields []Field
}

func NewMudEvent(table *Table, row []byte, fields []Field) MudEvent {
	keyAsString := hexutil.Encode(row)
	return MudEvent{
		Table:  table.Metadata.TableName,
		Key:    keyAsString,
		Fields: fields,
	}
}

func NewMudEventStringEncoded(table string, rowID string, fields []Field) MudEvent {
	return MudEvent{
		Table:  table,
		Key:    rowID,
		Fields: fields,
	}
}

func CreateUint32Event(table string, rowID string, value int64) MudEvent {
	return NewMudEventStringEncoded(table, rowID, []Field{
		{Key: "value", Data: NewUintFieldFromNumber(value)},
	})
}
