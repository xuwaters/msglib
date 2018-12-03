package msglib

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

const (
	MT_NULL   byte = 1
	MT_BOOL   byte = 2
	MT_BYTE   byte = 3
	MT_I16    byte = 4
	MT_I32    byte = 5
	MT_I64    byte = 6
	MT_FLOAT  byte = 7
	MT_DOUBLE byte = 8
	MT_BINARY byte = 9
	MT_STRING byte = 10
	MT_STRUCT byte = 11
	MT_MAP    byte = 12
	MT_LIST   byte = 13
	MT_SET    byte = 14
)

type MField struct {
	Name string
	Type byte
	ID   int
}

type MList struct {
	ElementType byte
	Count       int
}

type MMap struct {
	KeyType   byte
	ValueType byte
	Count     int
}

type MSet struct {
	ElementType byte
	Count       int
}

type MStruct struct {
	Name string
}

type IMProto interface {
	ReadStructBegin(reader io.Reader) (*MStruct, error)

	ReadFieldBegin(reader io.Reader) (*MField, error)

	ReadMapBegin(reader io.Reader) (*MMap, error)

	ReadListBegin(reader io.Reader) (*MList, error)

	ReadSetBegin(reader io.Reader) (*MSet, error)

	ReadBool(reader io.Reader) (bool, error)

	ReadByte(reader io.Reader) (byte, error)

	ReadI16(reader io.Reader) (int16, error)

	ReadI32(reader io.Reader) (int32, error)

	ReadI64(reader io.Reader) (int64, error)

	ReadFloat32(reader io.Reader) (float32, error)

	ReadFloat64(reader io.Reader) (float64, error)

	ReadBinary(reader io.Reader) ([]byte, error)

	ReadString(reader io.Reader) (string, error)

	WriteStructBegin(writer io.Writer, marker *MStruct) error

	WriteFieldBegin(writer io.Writer, marker *MField) error

	WriteFieldStop(writer io.Writer) error

	WriteMapBegin(writer io.Writer, marker *MMap) error

	WriteListBegin(writer io.Writer, marker *MList) error

	WriteSetBegin(writer io.Writer, marker *MSet) error

	WriteBool(writer io.Writer, data bool) error

	WriteByte(writer io.Writer, data byte) error

	WriteI16(writer io.Writer, data int16) error

	WriteI32(writer io.Writer, data int32) error

	WriteI64(writer io.Writer, data int64) error

	WriteFloat32(writer io.Writer, data float32) error

	WriteFloat64(writer io.Writer, data float64) error

	WriteBinary(writer io.Writer, data []byte) error

	WriteString(writer io.Writer, str string) error
}

// implements IMProto interface
type mBinaryProto struct {
	writeBuffer []byte
	readBuffer  []byte
	oneByte     []byte
}

func NewBinaryProto() IMProto {
	proto := &mBinaryProto{
		writeBuffer: make([]byte, 32),
		readBuffer:  make([]byte, 32), // can be used to optimize short strings
	}
	return proto
}

type mByteReader struct {
	reader io.Reader
	buffer []byte
}

func newByteReader(reader io.Reader, buffer []byte) *mByteReader {
	if len(buffer) < 1 {
		buffer = make([]byte, 1)
	}
	return &mByteReader{
		reader: reader,
		buffer: buffer,
	}
}

func (r *mByteReader) ReadByte() (byte, error) {
	_, err := io.ReadFull(r.reader, r.buffer[:1])
	if err != nil {
		return byte(0), err
	}
	return r.buffer[0], nil
}

func (bin *mBinaryProto) readVarint(reader io.Reader) (int64, error) {
	if br, ok := reader.(io.ByteReader); ok {
		return binary.ReadVarint(br)
	}
	br := newByteReader(reader, bin.readBuffer)
	return binary.ReadVarint(br)
}

func (bin *mBinaryProto) readUvarint(reader io.Reader) (uint64, error) {
	if br, ok := reader.(io.ByteReader); ok {
		return binary.ReadUvarint(br)
	}
	br := newByteReader(reader, bin.readBuffer)
	return binary.ReadUvarint(br)
}

func (bin *mBinaryProto) writeVarint(writer io.Writer, val int64) error {
	n := binary.PutVarint(bin.writeBuffer, val)
	_, err := writer.Write(bin.writeBuffer[:n])
	return err
}

func (bin *mBinaryProto) writeUvarint(writer io.Writer, val uint64) error {
	n := binary.PutUvarint(bin.writeBuffer, val)
	_, err := writer.Write(bin.writeBuffer[:n])
	return err
}

func (bin *mBinaryProto) ReadStructBegin(reader io.Reader) (*MStruct, error) {
	return nil, nil
}

func (bin *mBinaryProto) WriteStructBegin(writer io.Writer, marker *MStruct) error {
	return nil
}

func (bin *mBinaryProto) ReadFieldBegin(reader io.Reader) (field *MField, err error) {
	val, err := bin.readUvarint(reader)
	if err != nil {
		return
	}
	idAndType := int32(val)
	field = &MField{}
	field.Type = byte(idAndType & 0x0F)
	field.ID = int(idAndType >> 4)
	return
}

func (bin *mBinaryProto) WriteFieldBegin(writer io.Writer, marker *MField) (err error) {
	idAndType := ((int32(marker.ID) << 4) | (int32(marker.Type) & 0x0F))
	err = bin.writeUvarint(writer, uint64(idAndType))
	return
}

func (bin *mBinaryProto) WriteFieldStop(writer io.Writer) error {
	marker := &MField{Name: "", Type: MT_NULL, ID: 0}
	return bin.WriteFieldBegin(writer, marker)
}

func (bin *mBinaryProto) ReadMapBegin(reader io.Reader) (*MMap, error) {
	val, err := bin.readUvarint(reader)
	if err != nil {
		return nil, err
	}
	marker := &MMap{}
	marker.Count = int(val)
	tval, err := bin.ReadByte(reader)
	if err != nil {
		return nil, err
	}
	marker.KeyType = byte(tval & 0x0F)
	marker.ValueType = byte(tval >> 4)
	return marker, nil
}

func (bin *mBinaryProto) WriteMapBegin(writer io.Writer, marker *MMap) error {
	err := bin.writeUvarint(writer, uint64(marker.Count))
	if err != nil {
		return err
	}
	key := byte(marker.KeyType)
	val := byte(marker.ValueType)
	tval := byte(((val & 0x0F) << 4) | (key & 0x0F))
	err = bin.WriteByte(writer, tval)
	return err
}

func (bin *mBinaryProto) ReadListBegin(reader io.Reader) (*MList, error) {
	val, err := bin.readUvarint(reader)
	if err != nil {
		return nil, err
	}
	countAndType := int32(val)
	list := &MList{}
	list.ElementType = byte(countAndType & 0x0F)
	list.Count = int(countAndType >> 4)
	return list, nil
}

func (bin *mBinaryProto) WriteListBegin(writer io.Writer, marker *MList) error {
	countAndType := uint64((marker.Count << 4) | int(marker.ElementType&0x0F))
	err := bin.writeUvarint(writer, countAndType)
	return err
}

func (bin *mBinaryProto) ReadSetBegin(reader io.Reader) (*MSet, error) {
	list, err := bin.ReadListBegin(reader)
	if err != nil {
		return nil, err
	}
	set := &MSet{ElementType: list.ElementType, Count: list.Count}
	return set, err
}

func (bin *mBinaryProto) WriteSetBegin(writer io.Writer, marker *MSet) error {
	list := &MList{}
	list.Count = marker.Count
	list.ElementType = marker.ElementType
	return bin.WriteListBegin(writer, list)
}

func (bin *mBinaryProto) ReadBool(reader io.Reader) (bool, error) {
	val, err := bin.ReadByte(reader)
	return (val == 1), err
}

func (bin *mBinaryProto) WriteBool(writer io.Writer, data bool) error {
	var val byte
	if data {
		val = byte(1)
	} else {
		val = byte(0)
	}
	return bin.WriteByte(writer, val)
}

func (bin *mBinaryProto) ReadByte(reader io.Reader) (byte, error) {
	onebyte := bin.readBuffer[:1]
	_, err := io.ReadFull(reader, onebyte)
	return onebyte[0], err
}

func (bin *mBinaryProto) WriteByte(writer io.Writer, data byte) error {
	onebyte := bin.writeBuffer[:1]
	onebyte[0] = data
	_, err := writer.Write(onebyte)
	return err
}

func (bin *mBinaryProto) ReadI16(reader io.Reader) (int16, error) {
	val, err := bin.readVarint(reader)
	return int16(val), err
}

func (bin *mBinaryProto) WriteI16(writer io.Writer, data int16) error {
	return bin.writeVarint(writer, int64(data))
}

func (bin *mBinaryProto) ReadI32(reader io.Reader) (int32, error) {
	val, err := bin.readVarint(reader)
	return int32(val), err
}

func (bin *mBinaryProto) WriteI32(writer io.Writer, data int32) error {
	return bin.writeVarint(writer, int64(data))
}

func (bin *mBinaryProto) ReadI64(reader io.Reader) (int64, error) {
	val, err := bin.readVarint(reader)
	return val, err
}

func (bin *mBinaryProto) WriteI64(writer io.Writer, data int64) error {
	return bin.writeVarint(writer, data)
}

func (bin *mBinaryProto) ReadFloat32(reader io.Reader) (float32, error) {
	buf := bin.readBuffer[0:4]
	_, err := io.ReadFull(reader, buf)
	val := math.Float32frombits(binary.LittleEndian.Uint32(buf))
	return val, err
}

func (bin *mBinaryProto) WriteFloat32(writer io.Writer, data float32) error {
	buf := bin.writeBuffer[0:4]
	binary.LittleEndian.PutUint32(buf, math.Float32bits(data))
	_, err := writer.Write(buf)
	return err
}

func (bin *mBinaryProto) ReadFloat64(reader io.Reader) (float64, error) {
	buf := bin.readBuffer[0:8]
	_, err := io.ReadFull(reader, buf)
	val := math.Float64frombits(binary.LittleEndian.Uint64(buf))
	return val, err
}

func (bin *mBinaryProto) WriteFloat64(writer io.Writer, data float64) error {
	buf := bin.writeBuffer[0:8]
	binary.LittleEndian.PutUint64(buf, math.Float64bits(data))
	_, err := writer.Write(buf)
	return err
}

func (bin *mBinaryProto) ReadBinary(reader io.Reader) ([]byte, error) {
	cnt, err := bin.readUvarint(reader)
	if err != nil || cnt == 0 {
		return nil, err
	} else if cnt < 0 {
		return nil, errors.New("negative length in msglib.BinaryProto.ReadBinary")
	}
	buf := make([]byte, cnt)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (bin *mBinaryProto) WriteBinary(writer io.Writer, binary []byte) error {
	cnt := len(binary)
	if err := bin.writeUvarint(writer, uint64(cnt)); err != nil {
		return err
	}
	_, err := writer.Write(binary)
	return err
}

func (bin *mBinaryProto) ReadString(reader io.Reader) (string, error) {
	cnt, err := bin.readUvarint(reader)
	if err != nil || cnt == 0 {
		return "", err
	} else if cnt < 0 {
		return "", errors.New("negative length in msglib.BinaryProto.ReadBinary")
	}
	// use readbuffer to optimize
	buf := bin.readBuffer
	if int(cnt) > len(buf) {
		buf = make([]byte, cnt)
	} else {
		buf = buf[:cnt]
	}
	if _, err := io.ReadFull(reader, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func (bin *mBinaryProto) WriteString(writer io.Writer, str string) error {
	buffer := []byte(str)
	return bin.WriteBinary(writer, buffer)
}

type UnsupportedValueError struct {
	Value   reflect.Value
	Message string
}

func (e *UnsupportedValueError) Error() string {
	return fmt.Sprintf("msglib: unsupported value (%+v): %s", e.Value, e.Message)
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return fmt.Sprintf("msglib: unsupported type: %+v", e.Type)
}
