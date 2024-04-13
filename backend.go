package main

import (
	"fmt"
	"sort"
	"strings"
)

type Backend struct {
	storage         Storage
	currentRow      Row
	tableDefinition TableDefinition
	functionCalls   map[string]*FunctionCall
	functionsData   map[string]map[string]*FunctionData
}

type SelectRow struct {
	Items   []interface{}
	OrderBy interface{}
}

type FunctionData struct {
	Function *FunctionCall
	Acc      interface{}
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

	backend.tableDefinition, _ = backend.storage.GetTableDefinition(statement.Table)

	items := expandSelectItems(*statement.Items, backend.tableDefinition)

	// Determine aggregate functions
	backend.functionCalls = make(map[string]*FunctionCall)
	backend.functionsData = make(map[string]map[string]*FunctionData)
	for _, item := range items {
		if item.Kind == FunctionCallExpressionKind {
			backend.functionCalls[item.FunctionCall.Name] = &item.FunctionCall
		}
	}

	// Should group data if group by is specified or if statement contains
	// aggregate functions (currently, all functions are aggregate functions)
	grouping := statement.GroupBy != (Expression{}) || len(backend.functionCalls) > 0
	if grouping {
		groupedData = make(map[string]*SelectRow)
	}

	// Sequential scan through table rows
	for index, row := range backend.storage.TableRows(statement.Table) {
		backend.currentRow = row
		// Break loop after reaching limit
		if statement.Limit != -1 &&
			index >= statement.Limit &&
			!grouping &&
			statement.OrderBy == (OrderBy{}) {
			break
		}
		// Apply where condition
		if statement.Where != (Expression{}) {
			if backend.evaluateExpression(statement.Where, "") != true {
				continue
			}
		}
		// Add row into groupedData or resultSet
		var groupKey string
		selectRow := new(SelectRow)
		if grouping {
			if statement.GroupBy != (Expression{}) {
				groupKey = interfaceToString(backend.evaluateExpression(statement.GroupBy, groupKey))
			}
			groupedData[groupKey] = selectRow
			// Process aggregate functions
			if backend.functionsData[groupKey] == nil {
				backend.functionsData[groupKey] = make(map[string]*FunctionData)
			}
			for _, function := range backend.functionCalls {
				fdata := backend.functionsData[groupKey][function.Name]
				switch function.Name {
				case "count":
					if fdata == nil {
						fdata = &FunctionData{Acc: 0}
						backend.functionsData[groupKey][function.Name] = fdata
					}
					fdata.Acc = fdata.Acc.(int) + 1
				}
			}
		} else {
			resultSet = append(resultSet, selectRow)
		}
		// Select items from row
		for _, item := range items {
			selectRow.Items = append(selectRow.Items, backend.evaluateExpression(item, groupKey))
		}
		// Evaluate and store value for order by
		if statement.OrderBy != (OrderBy{}) {
			selectRow.OrderBy = backend.evaluateExpression(statement.OrderBy.By, groupKey)
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

func (backend Backend) evaluateExpression(expression Expression, groupKey string) interface{} {
	switch expression.Kind {
	case IdentifierExpressionKind:
		return backend.currentRow.Values[backend.tableDefinition.ColumnIndexes[expression.Identifier]].Value
	case FunctionCallExpressionKind:
		return backend.functionsData[groupKey][expression.FunctionCall.Name].Acc
	case BinaryExpressionKind:
		a := backend.evaluateExpression(expression.Binary.A, groupKey)
		b := backend.evaluateExpression(expression.Binary.B, groupKey)
		switch expression.Binary.Operator {
		case "=":
			return a == b
		case "<>":
			return a != b
		case ">":
			return evaluateAGtB(a, b)
		case ">=":
			return evaluateAGteB(a, b)
		case "<":
			return evaluateALtB(a, b)
		case "<=":
			return evaluateALteB(a, b)
		}
	case LiteralExpressionKind:
		return expression.Literal
	}
	return "?"
}
