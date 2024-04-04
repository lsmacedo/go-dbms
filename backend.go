package main

import (
	"fmt"
	"sort"
	"strings"
)

type Backend struct {
	tables map[string]table
}

type table struct {
	columns *[]ColumnDefinition
	data    *[]byte
}

type SelectedRow struct {
	Data  []string
	Order string
}

type AggregateFunction struct {
	ItemIndex int
	Function  Function
	Data      map[string]int
}

func NewBackend() *Backend {
	return &Backend{
		tables: map[string]table{},
	}
}

func (backend *Backend) Run(statement Statement) {
	var returnedData [][]string
	switch statement.Kind {
	case CreateTableKind:
		backend.runCreateTable(statement.CreateTable)
	case InsertKind:
		backend.runInsert(statement.Insert)
	case SelectKind:
		returnedData = backend.runSelect(statement.Select)
	}

	for i := range returnedData {
		fmt.Println(strings.Join(returnedData[i], ", "))
	}
	if returnedData != nil {
		fmt.Println()
	}
}

func (backend *Backend) runCreateTable(statement CreateTableStatement) {
	backend.tables[statement.Name] = table{
		columns: statement.Columns,
		data:    &[]byte{},
	}
}

func (backend *Backend) runInsert(statement InsertStatement) {
	table := backend.tables[statement.Table]

	var rowData []byte
	for i := 0; i < len(*table.columns); i++ {
		var value *string
		for j := range *statement.Columns {
			if (*table.columns)[i].Name == (*statement.Columns)[j].Identifier {
				value = &(*statement.Values)[j].Literal
				break
			}
		}
		rowData = append(rowData, valueToByteArray(value, (*table.columns)[i].Type)...)
	}

	// Row Structure
	// 4 bytes - next row offset
	// n bytes - row data
	*table.data = append(*table.data, intToByteArray(INT_SIZE+len(rowData))...)
	*table.data = append(*table.data, rowData...)
}

func (backend Backend) runSelect(statement SelectStatement) [][]string {
	var resultSet []*SelectedRow
	var groupedData map[string]*SelectedRow

	table := backend.tables[statement.Table]

	selectItems := expandSelectItems(statement, table)
	aggregateFunctions := determineAggregateFunctions(selectItems)

	// Should group data if group by is specified or if there are aggregate functions
	grouping := statement.GroupBy != (Expression{}) || len(aggregateFunctions) > 0
	if grouping {
		groupedData = make(map[string]*SelectedRow)
	}

	// Select rows
	for cursor := 0; cursor < len(*table.data); cursor += intFromByteArray(*table.data, cursor) {
		// Apply where condition
		if statement.Where != (Expression{}) {
			if eval(statement.Where, table, cursor, nil, nil) == "0" {
				continue
			}
		}

		row := &SelectedRow{}

		// Determine group key
		var groupKey string
		if statement.GroupBy != (Expression{}) {
			groupKey = eval(statement.GroupBy, table, cursor, nil, nil)
		}

		// Group data
		if grouping {
			groupedData[groupKey] = row

			// Evaluate aggregate functions
			for _, function := range aggregateFunctions {
				if function.Function.Name == "count" {
					function.Data[groupKey]++
				}
			}
		}

		// Append row into selected data array
		if !grouping {
			resultSet = append(resultSet, row)
		}

		// Select specified items from row
		for _, item := range selectItems {
			row.Data = append(row.Data, eval(item, table, cursor, &groupKey, aggregateFunctions))
		}

		// Store column to use for sorting
		if statement.OrderBy != (OrderByExpression{}) {
			row.Order = eval(statement.OrderBy.By, table, cursor, &groupKey, aggregateFunctions)
		}
	}

	// Add grouped data into result set
	if grouping {
		for key := range groupedData {
			resultSet = append(resultSet, groupedData[key])
		}
	}

	// Sort results
	if statement.OrderBy != (OrderByExpression{}) {
		sort.Slice(resultSet, func(i, j int) bool {
			if statement.OrderBy.Direction == "asc" {
				return resultSet[i].Order < resultSet[j].Order
			} else {
				return resultSet[i].Order > resultSet[j].Order
			}
		})
	}

	// Turn into [][]string
	var response [][]string
	for _, row := range resultSet {
		response = append(response, row.Data)
	}

	// Apply limit
	limit := statement.Limit
	if limit == -1 || limit > len(response) {
		limit = len(response)
	}

	// Apply offset
	offset := statement.Offset
	if statement.Offset == -1 {
		offset = 0
	}

	return response[offset:limit]
}
