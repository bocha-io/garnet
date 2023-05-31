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
