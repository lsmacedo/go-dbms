package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
)

const (
	TableDefinitionsIndex int = 0
	PageDirectoryIndex    int = 1
	DataStartIndex        int = 2
)

type NumericTypeSize uint

const (
	SmallIntSize NumericTypeSize = 2
	IntSize      NumericTypeSize = 4
)

type ColumnType uint

const (
	Int ColumnType = iota
	Text
	Unknown
)

type Storage struct {
	filePath string
	pageSize int
	// May in the future have properties to cache pages
}

type TableDefinition struct {
	Name          string
	Columns       []ColumnDefinition
	ColumnIndexes map[string]int
}

type Row struct {
	Values []RowValue
}

type RowValue struct {
	Column string
	Value  interface{}
}

func NewStorage() Storage {
	s := Storage{
		filePath: "data",
		pageSize: 16 * 1024,
	}
	if _, err := os.Stat(s.filePath); errors.Is(err, os.ErrNotExist) {
		// Create file if not exists
		os.Create(s.filePath)
		// Add two pages: table definitions and page directory
		s.createPage("table_definitions", false)
		s.createPage("page_directory", false)
	}
	return s
}

func (s Storage) CreateTable(tableName string, columns []ColumnDefinition) error {
	// Create buffer with column definitions
	cd := NewByteStreamBuffer()
	for _, column := range columns {
		cd.WriteString(column.Name)
		cd.WriteInt(int(columnTypeFromString(column.Type)), SmallIntSize)
	}

	// Create buffer to be written into page
	buf := NewByteStreamBuffer()
	buf.WriteString(tableName)
	buf.WriteInt(len(cd.Bytes()), IntSize)
	buf.Concat(cd)

	// Write into page
	err := s.appendToPage(buf.Bytes(), int(TableDefinitionsIndex))
	if err != nil {
		return err
	}

	return nil
}

func (s Storage) InsertInto(tableToInsert string, values []RowValue) error {
	var maxPageIndex = -1

	tableDefinition, err := s.GetTableDefinition(tableToInsert)

	// Write values from row into a buffer
	buf := NewByteStreamBuffer()
	for _, column := range tableDefinition.Columns {
		var value interface{}
		for i := range values {
			if values[i].Column == column.Name {
				value = values[i].Value
			}
		}
		switch column.Type {
		case "text":
			if value == nil {
				buf.WriteInt(math.MaxInt8, SmallIntSize)
			} else {
				buf.WriteString(value.(string))
			}
		case "integer":
			if value == nil {
				buf.WriteInt(math.MaxInt32, IntSize)
			} else {
				buf.WriteInt(value.(int), IntSize)
			}
		}

	}

	// Read page directory to find latest page containing data for this table
	pd, err := s.readPage(int(PageDirectoryIndex))
	if err != nil {
		return err
	}
	pageLength := pd.ReadInt(IntSize)
	for pd.Cursor() < pageLength {
		tableName := pd.ReadString()
		pageIndex := pd.ReadInt(SmallIntSize)
		if tableName == tableToInsert {
			maxPageIndex = pageIndex
		}
	}

	if maxPageIndex == -1 {
		// If a page for the table cannot be found, create a new one
		maxPageIndex, err = s.createPage(tableToInsert, true)
		if err != nil {
			return err
		}
	} else {
		// Check whether there is enough space on the page, and create a new one if
		// needed
		page, err := s.readPage(maxPageIndex)
		if err != nil {
			return err
		}
		if usedSpace := page.ReadInt(IntSize); s.pageSize-usedSpace < len(buf.Bytes()) {
			maxPageIndex, err = s.createPage(tableToInsert, true)
		}
	}

	// Write buffer into page
	err = s.appendToPage(buf.Bytes(), maxPageIndex)
	if err != nil {
		return err
	}

	return nil
}

func (s Storage) TableRows(tableName string) func(yield func(int, Row) bool) {
	var pages []int

	pd, _ := s.readPage(int(PageDirectoryIndex))

	// Get pages list
	pageLength := pd.ReadInt(IntSize)
	for pd.Cursor() < pageLength {
		table := pd.ReadString()
		pageIndex := pd.ReadInt(SmallIntSize)
		if table == tableName {
			pages = append(pages, pageIndex)
		}
	}

	// Get table definition
	tableDefinition, _ := s.GetTableDefinition(tableName)

	// This iterator will load pages as needed, and iterate through their rows
	return func(yield func(int, Row) bool) {
		var rowIndex int
		for _, pageIndex := range pages {
			page, _ := s.readPage(pageIndex)
			pageLength = page.ReadInt(IntSize)
			for page.Cursor() < pageLength {
				row := Row{}
				for _, column := range tableDefinition.Columns {
					var value interface{}
					switch column.Type {
					case "text":
						value = page.ReadString()
					case "integer":
						value = page.ReadInt(IntSize)
					}
					row.Values = append(row.Values, RowValue{Column: column.Name, Value: value})
				}
				if !yield(rowIndex, row) {
					return
				}
				rowIndex++
			}
		}
	}
}

func (s Storage) GetTableDefinition(tableName string) (TableDefinition, error) {
	var tableDefinition TableDefinition
	var tdFound bool
	var tdLength int

	buf, err := s.readPage(int(TableDefinitionsIndex))
	if err != nil {
		return tableDefinition, err
	}

	// Find definition of current table
	pageLength := buf.ReadInt(IntSize)
	for buf.Cursor() < pageLength {
		table := buf.ReadString()
		tdLength = buf.ReadInt(IntSize)
		if table == tableName {
			tdFound = true
			tableDefinition.Name = table
			break
		}
		buf.Skip(tdLength)
	}

	if !tdFound {
		return tableDefinition, fmt.Errorf("definition for table %s not found", tableName)
	}

	tableDefinition.ColumnIndexes = make(map[string]int)
	tdEnd := buf.Cursor() + tdLength
	i := 0
	for buf.Cursor() < tdEnd {
		column := buf.ReadString()
		tableDefinition.Columns = append(tableDefinition.Columns, ColumnDefinition{
			Name: column,
			Type: columnTypeToString(ColumnType(buf.ReadInt(SmallIntSize))),
		})
		tableDefinition.ColumnIndexes[column] = i
		i++
	}

	return tableDefinition, nil
}

func (s Storage) createPage(tableName string, addToPageDirectory bool) (int, error) {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	// Create buffer with page length prefix
	buf := NewByteStreamBuffer()
	buf.WriteInt(int(IntSize), IntSize)

	// Write bytes from buffer at the next available page location
	stat, err := file.Stat()
	if err != nil {
		return -1, err
	}
	pageIndex := int(int(stat.Size()) / s.pageSize)
	if stat.Size() > 0 {
		pageIndex++
	}

	// Write page length into new page
	file.WriteAt(buf.Bytes(), int64(pageIndex*s.pageSize))

	// Add to page directory
	if addToPageDirectory {
		buf = NewByteStreamBuffer()
		buf.WriteString(tableName)
		buf.WriteInt(pageIndex, SmallIntSize)
		err = s.appendToPage(buf.Bytes(), int(PageDirectoryIndex))
		if err != nil {
			return -1, err
		}
	}

	return pageIndex, nil
}

func (s Storage) appendToPage(bytes []byte, pageIndex int) error {
	file, err := os.OpenFile(s.filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read page length, so we can write after this position
	plBytes := make([]byte, 4)
	file.ReadAt(plBytes, int64(pageIndex*s.pageSize))
	pageLength := binary.BigEndian.Uint32(plBytes)
	if pageLength == 0 {
		pageLength += uint32(IntSize)
	}

	// Write page length at page's first position
	wb := NewByteStreamBuffer()
	wb.WriteInt(int(pageLength)+len(bytes), IntSize)

	file.WriteAt(wb.Bytes(), int64(pageIndex*s.pageSize))

	// Write content at the end of the page
	file.WriteAt(bytes, int64(pageIndex*s.pageSize)+int64(pageLength))

	return nil
}

func (s Storage) readPage(pageIndex int) (ByteStreamBuffer, error) {
	file, err := os.OpenFile(s.filePath, os.O_RDONLY, 0666)
	if err != nil {
		return ByteStreamBuffer{}, err
	}
	defer file.Close()

	pageBytes := make([]byte, s.pageSize)
	file.ReadAt(pageBytes, int64(s.pageSize*pageIndex))

	return NewByteStreamBufferFrom(pageBytes), nil
}

func columnTypeFromString(columnType string) ColumnType {
	switch columnType {
	case "integer":
		return Int
	case "text":
		return Text
	}
	return Unknown
}

func columnTypeToString(columnType ColumnType) string {
	switch columnType {
	case Int:
		return "integer"
	case Text:
		return "text"
	default:
		return "unknown"
	}
}
