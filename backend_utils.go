package main

import (
	"bytes"
	"encoding/binary"
	"strconv"
)

const INT_SIZE = 4
const SMALL_INT_SIZE = 2

func valueToByteArray(value *string, columnType string) []byte {
	var rowData = []byte{}
	switch columnType {
	case "text":
		// Structure:
		// 2 bytes - text length
		// n bytes - value
		if value == nil {
			return make([]byte, SMALL_INT_SIZE)
		}
		rowData = append(rowData, smallIntToByteArray(len(*value))...)
		rowData = append(rowData, []byte(*value)...)
	case "integer":
		// Structure:
		// 4 bytes - value
		if value == nil {
			return make([]byte, INT_SIZE)
		}
		intValue, err := strconv.Atoi(*value)
		if err != nil {
			panic("invalid integer value")
		}
		rowData = append(rowData, intToByteArray(intValue)...)
	}
	return rowData
}

func intToByteArray(i int) (arr []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int32(i))
	return buf.Bytes()
}

func smallIntToByteArray(i int) (arr []byte) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, int16(i))
	return buf.Bytes()
}

func intFromByteArray(byteArray []byte, cursor int) int {
	return int(binary.BigEndian.Uint32(byteArray[cursor : cursor+INT_SIZE]))
}

func smallIntFromByteArray(byteArray []byte, cursor int) int {
	return int(binary.BigEndian.Uint16(byteArray[cursor : cursor+SMALL_INT_SIZE]))
}

func expandSelectItems(statement SelectStatement, table table) []Expression {
	var selectItems []Expression
	for _, item := range *statement.Items {
		if item.Identifier == "*" {
			for i := range *table.columns {
				selectItems = append(selectItems, Expression{
					Kind:       IdentifierExpressionKind,
					Identifier: (*table.columns)[i].Name,
				})
			}
		} else {
			selectItems = append(selectItems, item)
		}
	}
	return selectItems
}

func determineAggregateFunctions(selectItems []Expression) []AggregateFunction {
	var aggregateFunctions []AggregateFunction
	for i := range selectItems {
		if selectItems[i].Kind == FunctionExpressionKind {
			aggregateFunctions = append(
				aggregateFunctions,
				AggregateFunction{
					ItemIndex: i,
					Function:  selectItems[i].Function,
					Data:      make(map[string]int),
				},
			)
		}
	}
	return aggregateFunctions
}

func eval(expression Expression, table table, cursor int, groupKey *string, functions []AggregateFunction) string {
	if expression == (Expression{}) {
		return "?"
	}
	switch expression.Kind {
	case FunctionExpressionKind:
		for i := range functions {
			if functions[i].Function.Name == expression.Function.Name {
				return strconv.Itoa(functions[i].Data[*groupKey])
			}
		}
	case IdentifierExpressionKind:
		var columnType string
		cursor += INT_SIZE // Skipping row header
		for _, column := range *table.columns {
			if column.Name == expression.Identifier {
				columnType = column.Type
				break
			}
			cursor += nextColumnCursor(column.Type, table, cursor)
		}
		return readColumnValue(columnType, table, cursor)
	case LiteralExpressionKind:
		return expression.Literal
	case BinaryExpressionKind:
		a := eval(expression.Binary.A, table, cursor, groupKey, functions)
		b := eval(expression.Binary.B, table, cursor, groupKey, functions)
		intA, _ := strconv.Atoi(a)
		intB, _ := strconv.Atoi(b)
		switch expression.Binary.Operator {
		case "=":
			return booleanToString(a == b)
		case "<>":
			return booleanToString(a != b)
		case ">":
			return booleanToString(intA > intB)
		case ">=":
			return booleanToString(intA >= intB)
		case "<":
			return booleanToString(intA < intB)
		case "<=":
			return booleanToString(intA <= intB)
		case "%":
			return strconv.Itoa(intA % intB)
		}
		return "?"
	}
	return "?"
}

func booleanToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func nextColumnCursor(columnType string, table table, cursor int) int {
	if columnType == "text" {
		return smallIntFromByteArray(*table.data, cursor) + SMALL_INT_SIZE
	} else if columnType == "integer" {
		return INT_SIZE
	}
	return 0
}

func readColumnValue(columnType string, table table, cursor int) string {
	if columnType == "text" {
		length := smallIntFromByteArray(*table.data, cursor)
		cursor += SMALL_INT_SIZE
		return string((*table.data)[cursor : cursor+length])
	} else if columnType == "integer" {
		return strconv.Itoa(intFromByteArray(*table.data, cursor))
	}
	return "?"
}
