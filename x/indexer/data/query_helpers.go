package data

import (
	"fmt"
	"strconv"
	"strings"
)

// Returns the field object, the key as string and error
func handleGetRowResult(tableName string, row []Field, err error) (Field, string, error) {
	if err != nil {
		return Field{}, "", fmt.Errorf("error getting the row from the table %s: %s", tableName, err.Error())
	}
	if len(row) > 0 {
		return row[0], strings.ReplaceAll(row[0].Data.String(), "\"", ""), nil
	}
	return Field{}, "", fmt.Errorf("row has no values")
}

func GetRowFromIDUsingBytes(db *Database, w *World, rowID [32]byte, tableName string) (Field, string, error) {
	table := w.GetTableByName(tableName)
	row, err := db.GetRowUsingBytes(table, rowID[:])
	return handleGetRowResult(tableName, row, err)
}

func GetRowFromIDUsingString(db *Database, w *World, rowID string, tableName string) (Field, string, error) {
	table := w.GetTableByName(tableName)
	row, err := db.GetRow(table, rowID)
	return handleGetRowResult(tableName, row, err)
}

func GetRowFieldsUsingString(db *Database, w *World, rowID string, tableName string) ([]Field, error) {
	table := w.GetTableByName(tableName)
	row, err := db.GetRow(table, rowID)
	if err != nil {
		return []Field{}, fmt.Errorf("error getting the row from the table %s: %s", tableName, err.Error())
	}
	return row, nil
}

func GetRowFieldsUsingBytes(db *Database, w *World, rowID [32]byte, tableName string) ([]Field, error) {
	table := w.GetTableByName(tableName)
	row, err := db.GetRowUsingBytes(table, rowID[:])
	if err != nil {
		return []Field{}, fmt.Errorf("error getting the row from the table %s: %s", tableName, err.Error())
	}
	return row, nil
}

func handleInt64Result(tableName string, row []Field, err error) (int64, error) {
	if err != nil {
		return 0, err
	}

	if len(row) != 1 {
		return 0, fmt.Errorf("row from table %s has no value", tableName)
	}

	value, err := strconv.ParseInt(row[0].Data.String(), 10, 32)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func GetInt64UsingString(db *Database, w *World, rowID string, tableName string) (int64, error) {
	row, err := GetRowFieldsUsingString(db, w, rowID, tableName)
	return handleInt64Result(tableName, row, err)
}

func GetInt64UsingBytes(db *Database, w *World, rowID [32]byte, tableName string) (int64, error) {
	row, err := GetRowFieldsUsingBytes(db, w, rowID, tableName)
	return handleInt64Result(tableName, row, err)
}

func GetRows(db *Database, w *World, tableName string) map[string][]Field {
	table := w.GetTableByName(tableName)
	rows := db.GetRows(table)
	return rows
}

func GetBoolFromTable(db *Database, w *World, rowID string, tableName string) bool {
	_, value, err := GetRowFromIDUsingString(db, w, rowID, tableName)
	if err != nil {
		return false
	}
	return value == "true"
}
