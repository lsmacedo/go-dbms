package main

import (
	"fmt"
	"sort"
	"strings"
)

type Backend struct {
	storage Storage
}

func NewBackend() *Backend {
	return &Backend{storage: NewStorage()}
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

func (backend Backend) runCreateTable(statement CreateTableStatement) {
	backend.storage.CreateTable(statement.Name, *statement.Columns)
}

func (backend *Backend) runInsert(statement InsertStatement) {
	var rows []RowValue
	for i := range *statement.Values {
		row := RowValue{
			Column: (*statement.Columns)[i].Identifier,
			Value:  (*statement.Values)[i].Literal,
		}
		rows = append(rows, row)
	}
	backend.storage.InsertInto(statement.Table, rows)
}

func (backend Backend) runSelect(statement SelectStatement) [][]string {
	var result []Row
	var groupedData map[string]Row

	tableDefinition, _ := backend.storage.GetTableDefinition(statement.Table)

	// Should group data if group by is specified
	grouping := statement.GroupBy != (Expression{})
	if grouping {
		groupedData = make(map[string]Row)
	}

	// Sequential scan through table rows
	for index, row := range backend.storage.TableRows(statement.Table) {
		// Break loop after reaching limit
		if statement.Limit != -1 &&
			index >= statement.Limit &&
			statement.GroupBy == (Expression{}) &&
			statement.OrderBy == (OrderBy{}) {
			break
		}
		// Apply where condition
		if statement.Where != (Expression{}) {
			if evaluateExpression(statement.Where, row, tableDefinition) != true {
				continue
			}
		}
		// Determine group key
		var groupKey string
		if statement.GroupBy != (Expression{}) {
			evaluated := evaluateExpression(statement.GroupBy, row, tableDefinition)
			groupKey = interfaceToString(evaluated)
		}
		// Group data
		if grouping {
			groupedData[groupKey] = row
		}
		// Append row into selected data array
		if !grouping {
			result = append(result, row)
		}
	}

	// Add grouped data into result set
	if grouping {
		for key := range groupedData {
			result = append(result, groupedData[key])
		}
	}
	// Sort results
	if statement.OrderBy != (OrderBy{}) {
		// TODO support sorting by other expression types
		sortColumnIndex := tableDefinition.ColumnIndexes[statement.OrderBy.By.Identifier]
		sort.Slice(result, func(i, j int) bool {
			if statement.OrderBy.Direction == "asc" {
				return evaluateALtB(
					result[i].Values[sortColumnIndex].Value,
					result[j].Values[sortColumnIndex].Value,
				)
			} else {
				return evaluateAGtB(
					result[i].Values[sortColumnIndex].Value,
					result[j].Values[sortColumnIndex].Value,
				)
			}
		})
	}
	// Select items from rows
	items := expandSelectItems(*statement.Items, tableDefinition)
	response := make([][]string, len(result))
	for i, row := range result {
		for _, item := range items {
			// TODO support selecting other expression types
			columnIndex := tableDefinition.ColumnIndexes[item.Identifier]
			response[i] = append(response[i], interfaceToString(row.Values[columnIndex].Value))
		}
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
