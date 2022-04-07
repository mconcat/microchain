package store

import (
	"reflect"
	"strconv"
	"strings"
	"errors"

	"github.com/emicklei/proto"
	"google.golang.org/protobuf/encoding/protowire"
)

// Protoutil provies a way to reuse and manipulate protobuf messages.


// there are so many protobuf packages, including gogo, canonical protobuf, etc..
// just declare a simple one here
type ProtoMessage interface {
	ProtoMessage()
}

type ProtoField struct {
	FieldNumber int32
	FieldType   int8
	FieldIsRep  bool
}

func Marshal[T ProtoValue](field ProtoField, v T) []byte {
	res := make([]byte, 0)

	return Append(field, res, v)	
}

func Append[T ProtoValue](field ProtoField, bz []byte, vt T) []byte {
	v := any(vt) // I'm not sure if this is a right way...
	switch field.FieldIsRep {
	case false:
		bz = protowire.AppendTag(bz, protowire.Number(field.FieldNumber), protowire.Type(field.FieldType))
		switch field.FieldType {
		case 0:
			bz = protowire.AppendVarint(bz, v.(uint64))
		case 1:
			bz = protowire.AppendFixed64(bz, v.(uint64))
		case 2:
			bz = protowire.AppendBytes(bz, v.([]byte))
		}
	case true:
		switch field.FieldType {
		case 0:
			bz = protowire.AppendTag(bz, protowire.Number(field.FieldNumber), protowire.Type(2))
			vbz := make([]byte, 0)
			for _, elem := range v.([]uint64) {
				vbz = protowire.AppendVarint(vbz, elem)
			}
			bz = protowire.AppendBytes(bz, vbz)
		case 1:
			bz = protowire.AppendTag(bz, protowire.Number(field.FieldNumber), protowire.Type(2))
			vbz := make([]byte, 0)
			for _, elem := range v.([]uint64) {
				vbz = protowire.AppendFixed64(vbz, elem)
			}
			bz = protowire.AppendBytes(bz, vbz)
		case 2:
			for _, elem := range v.([][]byte) {
				bz = protowire.AppendTag(bz, protowire.Number(field.FieldNumber), protowire.Type(field.FieldType))
				bz = protowire.AppendBytes(bz, elem)
			}
		}
	}

	return bz
}

func Unmarshal[T ProtoValue](field ProtoField, bz []byte) (res T, err error) {
	fieldnum, fieldty, size := protowire.ConsumeTag(bz)
	fieldtyExpected := field.FieldType
	if field.FieldIsRep { fieldtyExpected = 2 }
	if protowire.Type(fieldtyExpected) != fieldty { return errors.New("asdf") }
	if protowire.Number(field.FieldNumber) != fieldnum { return errors.New("qwer") }
	bz = bz[size:]

	var resv any

	switch field.FieldIsRep {
	case false:
		switch field.FieldType {
		case 0:
			resv, _ = protowire.ConsumeVarint(bz)
		case 1:
			resv, _ = protowire.ConsumeFixed64(bz)
		case 2:
			resv, _ = protowire.ConsumeBytes(bz)
		}
	case true:
		switch field.FieldType {
		case 0:
			values := []uint64{}
			repeated, _ := protowire.ConsumeBytes(bz)
			for len(repeated) > 0 {
				value, size := protowire.ConsumeVarint(repeated)
				values = append(values, value)
				repeated = repeated[size:]
			}
			resv = values
		case 1:
			values := []uint64{}
			repeated, _ := protowire.ConsumeBytes(bz)
			for len(repeated) > 0 {
				value, size := protowire.ConsumeFixed64(repeated)
				values = append(values, value)
				repeated = repeated[size:]
			}
			resv = values
		case 2:
			values := [][]byte{}
			repeated, _ := protowire.ConsumeBytes(bz)
			for len(repeated) > 0 {
				value, size := protowire.ConsumeBytes(repeated)
				values = append(values, value)
				repeated = repeated[size:]
			}
			resv = values
		}
	}

	return resv.(T), nil
}

func (field ProtoField) Search(bz []byte) (start, size int) {
	for len(bz) > 0 {
		fieldnum, _, size := protowire.ConsumeField(bz)
		if fieldnum == protowire.Number(field.FieldNumber) {
			return start, size
		}
		start = start + size
		bz = bz[size:]
	}
	return 0, 0
}

var (
	globalProtoFieldCache map[reflect.Type]map[string]ProtoField
)

func protoFields[T ProtoMessage](v T) map[string]ProtoField {
	ty := reflect.TypeOf(v)
	if m, ok := globalProtoFieldCache[ty]; ok {
		return m
	}

	m := make(map[string]ProtoField)
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)
		tags := strings.Split(field.Tag.Get("protobuf"), ",")
		fieldtystr, fieldnumstr, fieldrepstr := tags[0], tags[1], tags[2]
		var fieldty int8
		switch fieldtystr {
		case "varint":
			fieldty = 0
		case "bytes":
			fieldty = 2
		// case "fixed32bit":
		// fieldty = 5
		case "fixed64bit":
			fieldty = 1
		default:
			panic("unknown field type")
		}
		fieldnum, err := strconv.ParseInt(fieldnumstr, 10, 32)
		if err != nil {
			panic(err)
		}
		m[field.Name] = ProtoField{int32(fieldnum), int8(fieldty), fieldrepstr == "rep"}
	}

	globalProtoFieldCache[ty] = m
	return m
}

type ProtoValue interface {
	~uint64 | ~[]byte | ~[]uint64 | ~[][]byte
}

type ProtoFieldValue[T ProtoValue] interface {
	GetFieldNumber() int32
	GetFieldType() int8
	GetFieldIsRep() bool
	Get() T
	Set(T) // in-place modification on protobuf message
}

type ProtoFieldValueImpl[T ProtoValue] struct {
	buf *[]byte // TODO: this will cause dirty write on buffer(by multiple fieldvalues). make it a proper struct.
	start, size int
	field ProtoField
}

func (fieldval *ProtoFieldValueImpl[T]) GetFieldNumber() int32 { return fieldval.field.FieldNumber }
func (fieldval *ProtoFieldValueImpl[T]) GetFieldType() int8 { return fieldval.field.FieldType }
func (fieldval *ProtoFieldValueImpl[T]) GetFieldIsRep() bool { return fieldval.field.FieldIsRep }
func (fieldval *ProtoFieldValueImpl[T]) Get() T {
	if fieldval.size == 0 { // this could represent non-existing field. distinguish TODO
		fieldval.start, fieldval.size = fieldval.field.Search(*fieldval.buf)
		if fieldval.size == 0 {
			var res T
			return res
		}
	}
	res, err := Unmarshal[T](fieldval.field, (*fieldval.buf)[fieldval.start:fieldval.start+fieldval.size])
	if err != nil {
		var res T
		return res
	}
	return res
}
func (fieldval *ProtoFieldValueImpl[T]) Set(v T) {
	if fieldval.size == 0 { // this could represent non-existing field. distinguish TODO
		fieldval.start, fieldval.size = fieldval.field.Search(*fieldval.buf)
		if fieldval.size == 0 {
			*fieldval.buf = Append(fieldval.field, *fieldval.buf, v)
			return
		}
	}
	newbuf := make([]byte, 0, len(*fieldval.buf))
	newbuf = append(newbuf, (*fieldval.buf)[:fieldval.start]...)
	newbuf = Append(fieldval.field, newbuf, v)
	newbuf = append(newbuf, (*fieldval.buf)[fieldval.start+fieldval.size:]...)
	*fieldval.buf = newbuf
}

