package data

import (
	"fmt"
	"sync"
	"time"

	"github.com/bocha-io/garnet/internal/indexer/data/mudhelpers"
	"github.com/bocha-io/garnet/internal/logger"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Event struct {
	// TODO: add world here
	// World string `json:"world"`
	Table string `json:"table"`
	Row   string `json:"row"`
	Value string `json:"value"`
}

type TableMetadata struct {
	TableID          string
	TableName        string
	OnChainTableName string
	WorldAddress     string
}

type TableSchema struct {
	FieldNames  *[]string
	KeyNames    *[]string
	Schema      *mudhelpers.SchemaTypeKV
	NamedFields *map[string]mudhelpers.SchemaType
}

type Table struct {
	Metadata *TableMetadata
	Schema   *TableSchema
	Rows     *map[string][]Field
}

type World struct {
	Address string
	Tables  map[string]*Table
}

// TODO: add a cache layer here to avoid the loop
func (w *World) GetTableByName(tableName string) *Table {
	for tableID := range w.Tables {
		if w.Tables[tableID].Metadata != nil {
			if w.Tables[tableID].Metadata.TableName == tableName {
				return w.Tables[tableID]
			}
		}
	}
	return nil
}

func (w *World) GetTable(tableID string) *Table {
	if table, ok := w.Tables[tableID]; ok {
		return table
	}
	w.Tables[tableID] = &Table{
		Metadata: &TableMetadata{TableID: tableID, TableName: "", OnChainTableName: "", WorldAddress: w.Address},
		Schema:   &TableSchema{FieldNames: &[]string{}, KeyNames: &[]string{}, Schema: &mudhelpers.SchemaTypeKV{}, NamedFields: &map[string]mudhelpers.SchemaType{}},
		Rows:     &map[string][]Field{},
	}
	table := w.Tables[tableID]
	return table
}

type UnconfirmedTransaction struct {
	Txhash string
	Events []MudEvent
}

type Database struct {
	Worlds                  map[string]*World
	Events                  []Event
	LastUpdate              time.Time
	LastHeight              uint64
	ChainID                 string
	UnconfirmedTransactions []UnconfirmedTransaction
	txSentMutex             *sync.Mutex
}

func NewDatabase() *Database {
	return &Database{
		Worlds:     map[string]*World{},
		Events:     make([]Event, 0),
		LastUpdate: time.Now(),
		LastHeight: 0,
		ChainID:    "",
		// TODO: use a list instead of array
		UnconfirmedTransactions: []UnconfirmedTransaction{},
		txSentMutex:             &sync.Mutex{},
	}
}

func (db *Database) AddTxSent(tx UnconfirmedTransaction) {
	db.txSentMutex.Lock()
	defer db.txSentMutex.Unlock()
	db.UnconfirmedTransactions = append(db.UnconfirmedTransactions, tx)
}

func (db *Database) AddEvent(tableName string, key string, fields *[]Field) {
	value := ""
	if fields != nil {
		value = "{"
		for i, v := range *fields {
			value += v.String()
			if i != len(*fields)-1 {
				value += ","
			}
		}
		value += "}"
	}
	db.Events = append(db.Events, Event{Table: tableName, Row: key, Value: value})
	db.LastUpdate = time.Now()
}

func (db *Database) GetWorld(worldID string) *World {
	if world, ok := db.Worlds[worldID]; ok {
		return world
	}
	db.Worlds[worldID] = &World{Address: worldID, Tables: map[string]*Table{}}
	logger.LogInfo(fmt.Sprintf("new world registered %s", worldID))
	world := db.Worlds[worldID]
	return world
}

func (db *Database) GetTable(worldID string, tableID string) *Table {
	world := db.GetWorld(worldID)
	return world.GetTable(tableID)
}

func (db *Database) AddRow(table *Table, key []byte, fields *[]Field) MudEvent {
	// Use the database to add and remove info so we can broadcast events to subs
	keyAsString := hexutil.Encode(key)
	// TODO: add locks here
	(*table.Rows)[keyAsString] = *fields
	db.AddEvent(table.Metadata.TableName, keyAsString, fields)
	return NewMudEvent(table, key, *fields)
}

func (db *Database) SetField(table *Table, key []byte, event *mudhelpers.StorecoreStoreSetField) MudEvent {
	// TODO: add locks here

	// keyAsString := string(key)
	keyAsString := hexutil.Encode(key)
	fields, modified := BytesToFieldWithDefaults(event.Data, *table.Schema.Schema.Value, event.SchemaIndex, table.Schema.FieldNames)

	_, ok := (*table.Rows)[keyAsString]
	if ok {
		// Edit the row because it already exists
		for i := range (*table.Rows)[keyAsString] {
			if (*table.Rows)[keyAsString][i].Key == modified.Key {
				(*table.Rows)[keyAsString][i].Data = modified.Data
				break
			}
		}
	} else {
		// Create an empty row with defaults but the event index that uses event.Data
		(*table.Rows)[keyAsString] = *fields
	}

	db.AddEvent(table.Metadata.TableName, keyAsString, fields)
	return NewMudEvent(table, key, *fields)
}

func (db *Database) DeleteRow(table *Table, key []byte) MudEvent {
	keyAsString := hexutil.Encode(key)
	// TODO: add locks here
	delete((*table.Rows), keyAsString)
	db.AddEvent(table.Metadata.TableName, keyAsString, nil)
	return NewMudEvent(table, key, nil)
}

func (db *Database) GetRowUsingBytes(table *Table, key []byte) ([]Field, error) {
	keyAsString := hexutil.Encode(key)
	return db.GetRow(table, keyAsString)
}

func (db *Database) GetRow(table *Table, key string) ([]Field, error) {
	var fields []Field
	found := false
	// TODO: go from the lastest to the first one so we can break the for loop instead of looking for the most recent value
	for _, v := range db.UnconfirmedTransactions {
		for _, event := range v.Events {
			if event.Table == table.Metadata.TableName {
				if key == event.Key {
					fields = event.Fields
					found = true
				}
			}
		}
	}

	if found {
		return fields, nil
	}

	// Look for the value in the database
	v, ok := (*table.Rows)[key]
	if ok {
		return v, nil
	}

	return []Field{}, fmt.Errorf("key not found")
}

func (db *Database) GetRows(table *Table) map[string][]Field {
	// TODO: improve this because it's expensive
	ret := map[string][]Field{}
	for k, v := range *table.Rows {
		temp := [4]Field{}
		copy(temp[:], v)
		ret[k] = temp[:]
	}

	for _, v := range db.UnconfirmedTransactions {
		for _, event := range v.Events {
			if event.Table == table.Metadata.TableName {
				ret[event.Key] = event.Fields
			}
		}
	}
	return ret
}
