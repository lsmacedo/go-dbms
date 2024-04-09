package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ByteStreamBuffer struct {
	buffer *bytes.Buffer
	cursor int
}

func NewByteStreamBuffer() ByteStreamBuffer {
	return ByteStreamBuffer{buffer: new(bytes.Buffer)}
}

func NewByteStreamBufferFrom(value []byte) ByteStreamBuffer {
	return ByteStreamBuffer{buffer: bytes.NewBuffer(value)}
}

func (wb *ByteStreamBuffer) WriteInt(value int, length NumericTypeSize) error {
	switch length {
	case SmallIntSize:
		return binary.Write(wb.buffer, binary.BigEndian, int16(value))
	case IntSize:
		return binary.Write(wb.buffer, binary.BigEndian, int32(value))
	default:
		return fmt.Errorf("unsupported size")
	}
}

func (wb *ByteStreamBuffer) WriteString(value string) {
	wb.WriteInt(len(value), SmallIntSize)
	wb.buffer.Write([]byte(value))
}

func (wb *ByteStreamBuffer) ReadInt(length NumericTypeSize) int {
	var value int
	switch length {
	case SmallIntSize:
		value = int(binary.BigEndian.Uint16(wb.buffer.Bytes()[wb.cursor : wb.cursor+int(length)]))
	case IntSize:
		value = int(binary.BigEndian.Uint32(wb.buffer.Bytes()[wb.cursor : wb.cursor+int(length)]))
	}
	wb.cursor += int(length)
	return value
}

func (wb *ByteStreamBuffer) ReadString() string {
	length := wb.ReadInt(SmallIntSize)
	value := string(wb.buffer.Bytes()[wb.cursor : wb.cursor+length])
	wb.cursor += length
	return value
}

func (wb *ByteStreamBuffer) Clear() {
	wb.buffer.Reset()
	wb.cursor = 0
}

func (wb *ByteStreamBuffer) Bytes() []byte {
	return wb.buffer.Bytes()
}

func (wb *ByteStreamBuffer) Length() int {
	return wb.buffer.Len()
}

func (wb *ByteStreamBuffer) Cursor() int {
	return wb.cursor
}

func (wb *ByteStreamBuffer) Skip(offset int) {
	wb.cursor += offset
}
