package main

import (
	"fmt"
	"sort"
	"strings"
)

type Backend struct {
	storage Storage
}

type SelectRow struct {
	Items   []interface{}
	OrderBy interface{}
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

func (backend Backend) runInsert(statement InsertStatement) {
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
	var resultSet []*SelectRow
	var groupedData map[string]*SelectRow
	var response [][]string

	tableDefinition, _ := backend.storage.GetTableDefinition(statement.Table)

	items := expandSelectItems(*statement.Items, tableDefinition)

	// Should group data if group by is specified
	grouping := statement.GroupBy != (Expression{})
	if grouping {
		groupedData = make(map[string]*SelectRow)
	}

	// Sequential scan through table rows
	for index, row := range backend.storage.TableRows(statement.Table) {
		// Break loop after reaching limit
		if statement.Limit != -1 &&
			index >= statement.Limit &&
			!grouping &&
			statement.OrderBy == (OrderBy{}) {
			break
		}
		// Apply where condition
		if statement.Where != (Expression{}) {
			if evaluateExpression(statement.Where, row, tableDefinition) != true {
				continue
			}
		}
		// Add row into groupedData or resultSet
		selectRow := new(SelectRow)
		if grouping {
			groupKey := interfaceToString(evaluateExpression(statement.GroupBy, row, tableDefinition))
			groupedData[groupKey] = selectRow
		} else {
			resultSet = append(resultSet, selectRow)
		}
		// Select items from row
		for _, item := range items {
			selectRow.Items = append(selectRow.Items, evaluateExpression(item, row, tableDefinition))
		}
		// Evaluate and store value for order by
		if statement.OrderBy != (OrderBy{}) {
			selectRow.OrderBy = evaluateExpression(statement.OrderBy.By, row, tableDefinition)
		}
	}

	// Add grouped data into result set
	if grouping {
		for key := range groupedData {
			resultSet = append(resultSet, groupedData[key])
		}
	}
	// Sort results
	if statement.OrderBy != (OrderBy{}) {
		sort.Slice(resultSet, func(i, j int) bool {
			if statement.OrderBy.Direction == "asc" {
				return evaluateALtB(resultSet[i].OrderBy, resultSet[j].OrderBy)
			} else {
				return evaluateAGtB(resultSet[i].OrderBy, resultSet[j].OrderBy)
			}
		})
	}
	// Turn result set into [][]string response
	for _, row := range resultSet {
		rowItems := make([]string, len(row.Items))
		for i, item := range row.Items {
			rowItems[i] = interfaceToString(item)
		}
		response = append(response, rowItems)
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
