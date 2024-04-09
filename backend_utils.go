package main

import "strconv"

func expandSelectItems(items []Expression, table TableDefinition) []Expression {
	var selectItems []Expression
	for _, item := range items {
		if item.Identifier == "*" {
			for i := range table.Columns {
				selectItems = append(selectItems, Expression{
					Kind:       IdentifierExpressionKind,
					Identifier: table.Columns[i].Name,
				})
			}
		} else {
			selectItems = append(selectItems, item)
		}
	}
	return selectItems
}

func evaluateExpression(expression Expression, row Row, table TableDefinition) interface{} {
	switch expression.Kind {
	case IdentifierExpressionKind:
		return row.Values[table.ColumnIndexes[expression.Identifier]].Value
	case BinaryExpressionKind:
		a := evaluateExpression(expression.Binary.A, row, table)
		b := evaluateExpression(expression.Binary.B, row, table)
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

func evaluateAGtB(a interface{}, b interface{}) bool {
	switch a.(type) {
	case int:
		return a.(int) > b.(int)
	case string:
		return a.(string) > b.(string)
	}
	return false
}

func evaluateAGteB(a interface{}, b interface{}) bool {
	switch a.(type) {
	case int:
		return a.(int) >= b.(int)
	case string:
		return a.(string) >= b.(string)
	}
	return false
}

func evaluateALtB(a interface{}, b interface{}) bool {
	switch a.(type) {
	case int:
		return a.(int) < b.(int)
	case string:
		return a.(string) < b.(string)
	}
	return false
}

func evaluateALteB(a interface{}, b interface{}) bool {
	switch a.(type) {
	case int:
		return a.(int) <= b.(int)
	case string:
		return a.(string) <= b.(string)
	}
	return false
}

func interfaceToString(i interface{}) string {
	switch i.(type) {
	case string:
		return i.(string)
	case int:
		return strconv.Itoa(i.(int))
	}
	return "?"
}
