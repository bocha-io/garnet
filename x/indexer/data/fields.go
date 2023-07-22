package data

import (
	"fmt"
	"math/big"

	"github.com/bocha-io/garnet/x/indexer/data/mudhelpers"
	"github.com/bocha-io/logger"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type FieldData interface {
	String() string
	Type() string
}

type Field struct {
	Key  string
	Data FieldData
}

func (f Field) String() string {
	return fmt.Sprintf("\"%s\":%s", f.Key, f.Data.String())
}

func (f Field) Type() string {
	return f.Data.Type()
}

type BytesField struct {
	Data []byte
}

func NewBytesField(data []byte) BytesField {
	return BytesField{Data: data}
}

func (f BytesField) String() string {
	return fmt.Sprintf("\"%s\"", hexutil.Encode(f.Data))
}

func (f BytesField) Type() string {
	return "BytesField"
}

type StringField struct {
	Data string
}

func NewStringField(data []byte) StringField {
	return StringField{Data: string(data)}
}

func NewStringFieldFromValue(value string) StringField {
	return StringField{Data: value}
}

func (f StringField) String() string {
	return fmt.Sprintf("\"%s\"", f.Data)
}

func (f StringField) Type() string {
	return "StringField"
}

type ArrayField struct {
	Data []FieldData
}

func NewArrayField(size int) ArrayField {
	return ArrayField{
		Data: make([]FieldData, size),
	}
}

func (f ArrayField) String() string {
	// TODO: improve the string builder performance
	ret := "["
	for k, v := range f.Data {
		ret += v.String()
		if k != len(f.Data)-1 {
			ret += ","
		}
	}
	ret += "]"
	return ret
}

func (f ArrayField) Type() string {
	return "ArrayField"
}

type UintField struct {
	Data big.Int
}

func NewUintField(data []byte) UintField {
	return UintField{
		Data: *new(big.Int).SetBytes(data),
	}
}

func NewUintFieldFromNumber(value int64) UintField {
	return UintField{
		Data: *big.NewInt(value),
	}
}

func (f UintField) String() string {
	return f.Data.String()
}

func (f UintField) Type() string {
	return "UintField"
}

type IntField struct {
	Data big.Int
}

func NewIntField(data []byte) IntField {
	return IntField{Data: *new(big.Int).SetBytes(data)}
}

func NewIntFieldFromNumber(value int64) IntField {
	return IntField{
		Data: *big.NewInt(value),
	}
}

func (f IntField) String() string {
	return f.Data.String()
}

func (f IntField) Type() string {
	return "IntField"
}

type BoolField struct {
	Data bool
}

func NewBoolField(encoding byte) BoolField {
	return BoolField{Data: encoding == 1}
}

func NewBoolFromValue(value bool) BoolField {
	return BoolField{Data: value}
}

func (f BoolField) String() string {
	if f.Data {
		return "true"
	}
	return "false"
}

func (f BoolField) Type() string {
	return "BoolField"
}

type AddressField struct {
	Data common.Address
}

func NewAddressField(encoding []byte) AddressField {
	return AddressField{Data: common.BytesToAddress(encoding)}
}

func (f AddressField) String() string {
	return fmt.Sprintf("\"%s\"", f.Data.Hex())
}

func (f AddressField) Type() string {
	return "AddressField"
}

func FieldWithDefautValue(schemaType mudhelpers.SchemaType) FieldData {
	// UINT8 - UINT256 is the first range. We add one to the schema type to get the
	// number of bytes to read, since enums start from 0 and UINT8 is the first one.
	if schemaType >= mudhelpers.UINT8 && schemaType <= mudhelpers.UINT256 {
		return NewUintField([]byte{0})
	}

	// INT8 - INT256 is the second range. We subtract UINT256 from the schema type
	// to account for the first range and re-set the bytes count to start from 1.
	if schemaType >= mudhelpers.INT8 && schemaType <= mudhelpers.INT256 {
		return NewIntField([]byte{0})
	}

	// BYTES is the third range. We subtract INT256 from the schema type to account
	// for the previous ranges and re-set the bytes count to start from 1.
	if schemaType >= mudhelpers.BYTES1 && schemaType <= mudhelpers.BYTES32 {
		return NewBytesField([]byte{0})
	}

	// BOOL is a standalone schema type.
	if schemaType == mudhelpers.BOOL {
		return NewBoolField(0)
	}

	// ADDRESS is a standalone schema type.
	if schemaType == mudhelpers.ADDRESS {
		return NewAddressField([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	}

	// STRING is a standalone schema type.
	if schemaType == mudhelpers.STRING {
		return NewStringField([]byte{})
	}

	// BYTES
	if schemaType == mudhelpers.BYTES {
		return NewBytesField([]byte{})
	}

	// ARRAYs
	if schemaType >= mudhelpers.UINT8_ARRAY && schemaType <= mudhelpers.ADDRESS_ARRAY {
		return NewArrayField(0)
	}
	logger.LogError(fmt.Sprintf("Unknown static field type %s", schemaType.String()))
	return nil
}
