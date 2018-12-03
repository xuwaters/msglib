package msglib

import (
	"bytes"
	"io"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

func Serialize(data interface{}) (payload []byte, err error) {
	writer := &bytes.Buffer{}
	proto := NewBinaryProto()
	err = EncodeStruct(writer, proto, data)
	if err != nil {
		return
	}
	payload = writer.Bytes()
	return
}

func Deserialize(payload []byte, data interface{}) (err error) {
	reader := bytes.NewBuffer(payload)
	proto := NewBinaryProto()
	err = DecodeStruct(reader, proto, data)
	return
}

// encode

type encoder struct {
	writer io.Writer
	proto  IMProto
}

func EncodeStruct(w io.Writer, proto IMProto, data interface{}) (err error) {
	//
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	enc := &encoder{writer: w, proto: proto}
	vo := reflect.ValueOf(data)
	enc.writeStruct(vo)
	return nil
}

func (enc *encoder) error(err interface{}) {
	panic(err)
}

func (enc *encoder) writeStruct(val reflect.Value) {
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		enc.error(&UnsupportedValueError{Value: val, Message: "expect a struct"})
	}
	marker := &MStruct{}
	if err := enc.proto.WriteStructBegin(enc.writer, marker); err != nil {
		enc.error(err)
	}
	for _, ef := range encodeFields(val.Type()).fields {
		field := val.Type().Field(ef.i)
		fieldValue := val.Field(ef.i)

		if isEmptyValue(fieldValue) {
			continue
		}

		mfield := &MField{Name: field.Name, Type: ef.fieldType, ID: ef.id}
		if err := enc.proto.WriteFieldBegin(enc.writer, mfield); err != nil {
			enc.error(err)
		}
		enc.writeValue(fieldValue, ef.fieldType)
	}
	enc.proto.WriteFieldStop(enc.writer)
}

func (enc *encoder) writeValue(val reflect.Value, valtype byte) {
	kind := val.Kind()
	if kind == reflect.Ptr || kind == reflect.Interface {
		val = val.Elem()
		kind = val.Kind()
	}

	var err error = nil
	switch valtype {
	case MT_BOOL:
		err = enc.proto.WriteBool(enc.writer, val.Bool())
	case MT_BYTE:
		if kind == reflect.Uint8 {
			err = enc.proto.WriteByte(enc.writer, byte(val.Uint()))
		} else {
			err = enc.proto.WriteByte(enc.writer, byte(val.Int()))
		}
	case MT_I16:
		if kind == reflect.Uint16 {
			err = enc.proto.WriteI16(enc.writer, int16(val.Uint()))
		} else {
			err = enc.proto.WriteI16(enc.writer, int16(val.Int()))
		}
	case MT_I32:
		if kind == reflect.Uint32 {
			err = enc.proto.WriteI32(enc.writer, int32(val.Uint()))
		} else {
			err = enc.proto.WriteI32(enc.writer, int32(val.Int()))
		}
	case MT_I64:
		if kind == reflect.Uint64 {
			err = enc.proto.WriteI64(enc.writer, int64(val.Uint()))
		} else {
			err = enc.proto.WriteI64(enc.writer, int64(val.Int()))
		}
	case MT_FLOAT:
		err = enc.proto.WriteFloat32(enc.writer, float32(val.Float()))
	case MT_DOUBLE:
		err = enc.proto.WriteFloat64(enc.writer, val.Float())
	case MT_BINARY:
		err = enc.proto.WriteBinary(enc.writer, val.Bytes())
	case MT_STRING:
		err = enc.proto.WriteString(enc.writer, val.String())
	case MT_STRUCT:
		enc.writeStruct(val)
	case MT_MAP:
		keytype := val.Type().Key()
		valtype := val.Type().Elem()

		mmap := &MMap{}
		mmap.Count = val.Len()
		mmap.KeyType = fieldType(keytype)
		mmap.ValueType = fieldType(valtype)
		if er := enc.proto.WriteMapBegin(enc.writer, mmap); er != nil {
			enc.error(er)
		}
		for _, k := range val.MapKeys() {
			enc.writeValue(k, mmap.KeyType)
			enc.writeValue(val.MapIndex(k), mmap.ValueType)
		}
	case MT_LIST:
		elemtype := val.Type().Elem()
		if elemtype.Kind() == reflect.Uint8 {
			err = enc.proto.WriteBinary(enc.writer, val.Bytes())
		} else {
			mlist := &MList{}
			mlist.Count = val.Len()
			mlist.ElementType = fieldType(elemtype)
			if er := enc.proto.WriteListBegin(enc.writer, mlist); er != nil {
				enc.error(er)
			}
			for i := 0; i < mlist.Count; i++ {
				enc.writeValue(val.Index(i), mlist.ElementType)
			}
		}
	case MT_SET:
		if val.Type().Kind() == reflect.Slice {
			elemtype := val.Type().Elem()
			mset := &MSet{}
			mset.Count = val.Len()
			mset.ElementType = fieldType(elemtype)
			if er := enc.proto.WriteSetBegin(enc.writer, mset); er != nil {
				enc.error(er)
			}
			for i := 0; i < mset.Count; i++ {
				enc.writeValue(val.Index(i), mset.ElementType)
			}
		} else if val.Type().Kind() == reflect.Map {
			elemtype := val.Type().Key()
			valuetype := val.Type().Elem()
			mset := &MSet{}
			mset.ElementType = fieldType(elemtype)
			if valuetype.Kind() == reflect.Bool {
				mset.Count = 0
				for _, k := range val.MapKeys() {
					if val.MapIndex(k).Bool() {
						mset.Count++
					}
				}
				if er := enc.proto.WriteSetBegin(enc.writer, mset); er != nil {
					enc.error(er)
				}
				for _, k := range val.MapKeys() {
					if val.MapIndex(k).Bool() {
						enc.writeValue(val.MapIndex(k), mset.ElementType)
					}
				}
			} else {
				if er := enc.proto.WriteSetBegin(enc.writer, mset); er != nil {
					enc.error(er)
				}
				for _, k := range val.MapKeys() {
					enc.writeValue(k, mset.ElementType)
				}
			}
		} else {
			enc.error(&UnsupportedTypeError{Type: val.Type()})
		}
	default:
		enc.error(&UnsupportedTypeError{Type: val.Type()})
	}

	if err != nil {
		enc.error(err)
	}
}

// decode
type decoder struct {
	reader io.Reader
	proto  IMProto
}

func (dec *decoder) error(err interface{}) {
	panic(err)
}

func DecodeStruct(reader io.Reader, proto IMProto, val interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	dec := &decoder{reader: reader, proto: proto}
	vo := reflect.ValueOf(val)
	dec.readStruct(vo)
	return nil
}

func (dec *decoder) readStruct(val reflect.Value) {
	if val.Kind() != reflect.Ptr {
		dec.error(&UnsupportedValueError{Value: val, Message: "expect pointer to struct"})
	}
	if val.Elem().Kind() != reflect.Struct {
		dec.error(&UnsupportedValueError{Value: val, Message: "expect a struct"})
	}
	dec.readValue(MT_STRUCT, val.Elem())
}

func (dec *decoder) readValue(msgtype byte, rfval reflect.Value) {
	ret := rfval
	kind := rfval.Kind()
	if kind == reflect.Ptr {
		if rfval.IsNil() {
			rfval.Set(reflect.New(rfval.Type().Elem()))
		}
		ret = rfval.Elem()
		kind = ret.Kind()
	}
	//
	var err error = nil
	switch msgtype {
	case MT_BOOL:
		if val, err := dec.proto.ReadBool(dec.reader); err != nil {
			dec.error(err)
		} else {
			ret.SetBool(val)
		}
	case MT_BYTE:
		if val, err := dec.proto.ReadByte(dec.reader); err != nil {
			dec.error(err)
		} else {
			if kind == reflect.Uint8 {
				ret.SetUint(uint64(val))
			} else {
				ret.SetInt(int64(val))
			}
		}
	case MT_I16:
		if val, err := dec.proto.ReadI16(dec.reader); err != nil {
			dec.error(err)
		} else {
			if kind == reflect.Uint16 {
				ret.SetUint(uint64(val))
			} else {
				ret.SetInt(int64(val))
			}
		}
	case MT_I32:
		if val, err := dec.proto.ReadI32(dec.reader); err != nil {
			dec.error(err)
		} else {
			if kind == reflect.Uint32 {
				ret.SetUint(uint64(val))
			} else {
				ret.SetInt(int64(val))
			}
		}
	case MT_I64:
		if val, err := dec.proto.ReadI64(dec.reader); err != nil {
			dec.error(err)
		} else {
			if kind == reflect.Uint64 {
				ret.SetUint(uint64(val))
			} else {
				ret.SetInt(val)
			}
		}
	case MT_FLOAT:
		if val, err := dec.proto.ReadFloat32(dec.reader); err != nil {
			dec.error(err)
		} else {
			ret.SetFloat(float64(val))
		}
	case MT_DOUBLE:
		if val, err := dec.proto.ReadFloat64(dec.reader); err != nil {
			dec.error(err)
		} else {
			ret.SetFloat(val)
		}
	case MT_BINARY:
		elemtype := ret.Type().Elem()
		elemtypename := elemtype.Name()
		if kind == reflect.Slice && elemtype.Kind() == reflect.Uint8 && (elemtypename == "uint8" || elemtypename == "byte") {
			if val, err := dec.proto.ReadBinary(dec.reader); err != nil {
				dec.error(err)
			} else {
				ret.SetBytes(val)
			}
		} else {
			err = &UnsupportedValueError{Value: ret, Message: "expect a byte array"}
		}
	case MT_STRING:
		if kind == reflect.Slice {
			dec.readValue(MT_BINARY, ret)
		} else {
			if val, err := dec.proto.ReadString(dec.reader); err != nil {
				dec.error(err)
			} else {
				ret.SetString(val)
			}
		}

	case MT_STRUCT:
		if _, err := dec.proto.ReadStructBegin(dec.reader); err != nil {
			dec.error(err)
		}

		meta := encodeFields(ret.Type())
		for {
			mfield, err := dec.proto.ReadFieldBegin(dec.reader)
			if err != nil {
				dec.error(err)
			}
			if mfield.Type == MT_NULL {
				break
			}

			ef, ok := meta.fields[int(mfield.ID)]
			if !ok {
				SkipValue(dec.reader, dec.proto, mfield.Type)
			} else {
				fval := ret.Field(ef.i)
				if mfield.Type != ef.fieldType {
					msg := "type mismatch: " + ret.Type().Name() + ", field: " + ef.name
					dec.error(&UnsupportedValueError{Value: ret, Message: msg})
				} else {
					dec.readValue(mfield.Type, fval)
				}
			}
		}

	case MT_MAP:
		keytype := ret.Type().Key()
		valtype := ret.Type().Elem()
		mmap, err := dec.proto.ReadMapBegin(dec.reader)
		if err != nil {
			dec.error(err)
		}
		ret.Set(reflect.MakeMap(ret.Type()))
		for i := 0; i < mmap.Count; i++ {
			key := reflect.New(keytype).Elem()
			val := reflect.New(valtype).Elem()
			dec.readValue(mmap.KeyType, key)
			dec.readValue(mmap.ValueType, val)
			ret.SetMapIndex(key, val)
		}

	case MT_LIST:
		elemtype := ret.Type().Elem()
		mlist, err := dec.proto.ReadListBegin(dec.reader)
		if err != nil {
			dec.error(err)
		}
		for i := 0; i < mlist.Count; i++ {
			val := reflect.New(elemtype).Elem()
			dec.readValue(mlist.ElementType, val)
			ret.Set(reflect.Append(ret, val))
		}

	case MT_SET:
		rettype := ret.Type()
		if rettype.Kind() == reflect.Slice {
			elemtype := rettype.Elem()
			mset, err := dec.proto.ReadSetBegin(dec.reader)
			if err != nil {
				dec.error(err)
			}
			for i := 0; i < mset.Count; i++ {
				val := reflect.New(elemtype).Elem()
				dec.readValue(mset.ElementType, val)
				ret.Set(reflect.Append(ret, val))
			}
		} else if rettype.Kind() == reflect.Map {
			elemtype := rettype.Key()
			valtype := rettype.Elem()
			mset, err := dec.proto.ReadSetBegin(dec.reader)
			if err != nil {
				dec.error(err)
			}
			ret.Set(reflect.MakeMap(rettype))
			for i := 0; i < mset.Count; i++ {
				key := reflect.New(elemtype).Elem()
				dec.readValue(mset.ElementType, key)
				switch valtype.Kind() {
				case reflect.Bool:
					ret.SetMapIndex(key, reflect.ValueOf(true))
				default:
					ret.SetMapIndex(key, reflect.Zero(valtype))
				}
			}
		} else {
			dec.error(&UnsupportedTypeError{Type: ret.Type()})
		}

	default:
		dec.error(&UnsupportedTypeError{Type: ret.Type()})
	}
	if err != nil {
		dec.error(err)
	}
	return
}

// helpers

// meta analyze
type encodeField struct {
	i         int // field index in struct
	id        int // msglib field id for struct
	fieldType byte
	name      string
}

type structMeta struct {
	fields map[int]encodeField
}

var (
	typeCacheLock     sync.RWMutex
	encodeFieldsCache = make(map[reflect.Type]structMeta)
)

func encodeFields(t reflect.Type) structMeta {
	typeCacheLock.RLock()
	m, ok := encodeFieldsCache[t]
	typeCacheLock.RUnlock()
	if ok {
		return m
	}

	typeCacheLock.Lock()
	defer typeCacheLock.Unlock()

	m, ok = encodeFieldsCache[t]
	if ok {
		return m
	}

	fs := make(map[int]encodeField)
	m = structMeta{fields: fs}
	v := reflect.Zero(t)
	n := v.NumField()
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.PkgPath != "" {
			continue
		}
		if f.Anonymous {
			continue
		}
		tv := f.Tag.Get("msglib")
		if tv == "-" {
			continue
		}
		if tv != "" {
			var ef encodeField
			ef.i = i
			id, opts := parseTag(tv)
			ef.id = id
			ef.name = f.Name
			if opts.Contains("set") {
				ef.fieldType = MT_SET
			} else {
				ef.fieldType = fieldType(f.Type)
			}
			fs[ef.id] = ef
		}
	}
	encodeFieldsCache[t] = m
	return m
}

func fieldType(t reflect.Type) byte {
	switch t.Kind() {
	case reflect.Bool:
		return MT_BOOL
	case reflect.Int8, reflect.Uint8:
		return MT_BYTE
	case reflect.Int16, reflect.Uint16:
		return MT_I16
	case reflect.Int32, reflect.Uint32, reflect.Int, reflect.Uint:
		return MT_I32
	case reflect.Int64, reflect.Uint64:
		return MT_I64
	case reflect.Float32:
		return MT_FLOAT
	case reflect.Float64:
		return MT_DOUBLE
	case reflect.Map:
		vt := t.Elem()
		if vt.Kind() == reflect.Struct && vt.Name() == "" && vt.NumField() == 0 {
			return MT_SET
		}
		return MT_MAP
	case reflect.Slice:
		et := t.Elem()
		if et.Kind() == reflect.Uint8 {
			return MT_BINARY
		} else {
			return MT_LIST
		}
	case reflect.Struct:
		return MT_STRUCT
	case reflect.String:
		return MT_STRING
	case reflect.Interface, reflect.Ptr:
		return fieldType(t.Elem())
	}
	panic(&UnsupportedTypeError{Type: t})
}

// tags

type tagOptions string

func parseTag(tag string) (int, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		id, _ := strconv.Atoi(tag[:idx])
		return id, tagOptions(tag[idx+1:])
	}
	id, _ := strconv.Atoi(tag)
	return id, tagOptions("")
}

func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.String:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	}
	return false
}

func SkipValue(reader io.Reader, proto IMProto, msgtype byte) error {
	var err error
	switch msgtype {
	case MT_BOOL:
		_, err = proto.ReadBool(reader)
	case MT_BYTE:
		_, err = proto.ReadByte(reader)
	case MT_I16:
		_, err = proto.ReadI16(reader)
	case MT_I32:
		_, err = proto.ReadI32(reader)
	case MT_I64:
		_, err = proto.ReadI64(reader)
	case MT_FLOAT:
		_, err = proto.ReadFloat32(reader)
	case MT_DOUBLE:
		_, err = proto.ReadFloat64(reader)
	case MT_BINARY, MT_STRING:
		_, err = proto.ReadBinary(reader)
	case MT_STRUCT:
		if _, err := proto.ReadStructBegin(reader); err != nil {
			return err
		}
		for {
			mfield, err := proto.ReadFieldBegin(reader)
			if err != nil {
				return err
			}
			if mfield.Type == MT_NULL {
				break
			}
			if err = SkipValue(reader, proto, mfield.Type); err != nil {
				return err
			}
		}
	case MT_MAP:
		mmap, err := proto.ReadMapBegin(reader)
		if err != nil {
			return err
		}
		for i := 0; i < mmap.Count; i++ {
			if err = SkipValue(reader, proto, mmap.KeyType); err != nil {
				return err
			}
			if err = SkipValue(reader, proto, mmap.ValueType); err != nil {
				return err
			}
		}
	case MT_LIST:
		mlist, err := proto.ReadListBegin(reader)
		if err != nil {
			return err
		}
		for i := 0; i < mlist.Count; i++ {
			if err = SkipValue(reader, proto, mlist.ElementType); err != nil {
				return err
			}
		}
	case MT_SET:
		mset, err := proto.ReadSetBegin(reader)
		if err != nil {
			return err
		}
		for i := 0; i < mset.Count; i++ {
			if err = SkipValue(reader, proto, mset.ElementType); err != nil {
				return err
			}
		}
	}
	return err
}
