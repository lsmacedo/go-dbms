package main

import (
	"errors"
	"fmt"
	"strings"
)

type Parser struct {
	tokens []Token
	cursor int
}

func NewParser() Parser {
	return Parser{}
}

func (p *Parser) Parse(tokens []Token) (Statement, error) {
	var emptyStatement Statement

	p.tokens = tokens
	p.cursor = 0

	// Look for create table statement
	createTableStatement, err := p.parseCreateTable()
	if err != nil {
		return emptyStatement, err
	}
	if createTableStatement != (CreateTableStatement{}) {
		return Statement{
			CreateTable: createTableStatement,
			Kind:        CreateTableKind,
		}, nil
	}

	// Look for insert statement
	insertStatement, err := p.parseInsert()
	if err != nil {
		return emptyStatement, err
	}
	if insertStatement != (InsertStatement{}) {
		return Statement{
			Insert: insertStatement,
			Kind:   InsertKind,
		}, nil
	}

	// Look for select statement
	selectStatement, err := p.parseSelect()
	if err != nil {
		return emptyStatement, err
	}
	if selectStatement != (SelectStatement{}) {
		return Statement{
			Select: selectStatement,
			Kind:   SelectKind,
		}, nil
	}

	return emptyStatement, errors.New("unable to identify operation type")
}

func (p *Parser) parseCreateTable() (CreateTableStatement, error) {
	var emptyStatement CreateTableStatement

	if !p.matchKeyword("create table") {
		return emptyStatement, nil
	}

	table := p.matchToken(Identifier)
	if table == (Token{}) {
		return emptyStatement, errors.New("expected identifier after 'create table'")
	}

	columns, err := p.parseCreateTableColumns()
	if err != nil {
		return emptyStatement, err
	}

	return CreateTableStatement{
		Name:    table.Value.(string),
		Columns: &columns,
	}, nil
}

func (p *Parser) parseCreateTableColumns() ([]ColumnDefinition, error) {
	var columns []ColumnDefinition

	if lp := p.matchToken(LeftParenthesis); lp == (Token{}) {
		return columns, errors.New("expected column definitions after 'create table'")
	}

	for {
		if p.matchToken(RightParenthesis) != (Token{}) {
			break
		}

		columnName := p.matchToken(Identifier)
		if columnName == (Token{}) {
			return columns, errors.New("expected column name")
		}

		columnType := p.matchToken(Keyword)
		if columnType == (Token{}) {
			return columns, fmt.Errorf("expected column type after '%s'", columnName.Value)
		}

		columns = append(
			columns,
			ColumnDefinition{Name: columnName.Value.(string), Type: columnType.Value.(string)},
		)

		p.matchToken(Comma)
	}
	return columns, nil
}

func (p *Parser) parseInsert() (InsertStatement, error) {
	var emptyStatement InsertStatement

	if !p.matchKeyword("insert into") {
		return emptyStatement, nil
	}

	table := p.matchToken(Identifier)
	if table == (Token{}) {
		return emptyStatement, errors.New("expected identifier after 'insert into'")
	}

	columns, err := p.parseInsertColumns()
	if err != nil {
		return emptyStatement, err
	}

	values, err := p.parseInsertValues()
	if err != nil {
		return emptyStatement, err
	}

	return InsertStatement{
		Table:   table.Value.(string),
		Columns: &columns,
		Values:  &values,
	}, nil
}

func (p *Parser) parseInsertColumns() ([]Expression, error) {
	var columns []Expression

	if lp := p.matchToken(LeftParenthesis); lp == (Token{}) {
		return columns, errors.New("expected columns list after 'insert into <table_name>'")
	}

	for {
		if p.matchToken(RightParenthesis) != (Token{}) {
			break
		}

		column := p.matchToken(Identifier)
		if column == (Token{}) {
			return columns, errors.New("expected column name")
		}

		columns = append(
			columns,
			Expression{Kind: IdentifierExpressionKind, Identifier: column.Value.(string)},
		)

		p.matchToken(Comma)
	}

	return columns, nil
}

func (p *Parser) parseInsertValues() ([]Expression, error) {
	var values []Expression
	if !p.matchKeyword("values") {
		return values, errors.New("expected 'values' after columns list")
	}

	if lp := p.matchToken(LeftParenthesis); lp == (Token{}) {
		return values, errors.New("expected values list after 'values'")
	}

	for {
		if p.matchToken(RightParenthesis) != (Token{}) {
			break
		}

		value := p.matchToken(Number, String)
		if value == (Token{}) {
			return values, errors.New("expected literal")
		}

		values = append(
			values,
			Expression{Kind: LiteralExpressionKind, Literal: value.Value},
		)

		p.matchToken(Comma)
	}

	return values, nil
}

func (p *Parser) parseSelect() (SelectStatement, error) {
	var emptyStatement SelectStatement

	if !p.matchKeyword("select") {
		return emptyStatement, nil
	}

	// Select ...
	items := p.parseSelectItems()

	// From ...
	table, err := p.parseSelectTable()
	if err != nil {
		return emptyStatement, err
	}

	// Where ...
	where, err := p.parseExpression("where")
	if err != nil {
		return emptyStatement, err
	}

	// Group by ...
	groupBy, err := p.parseExpression("group by")
	if err != nil {
		return emptyStatement, err
	}

	// Order by ...
	orderBy, err := p.parseOrderBy()
	if err != nil {
		return emptyStatement, err
	}

	// Limit ...
	limit, err := p.parseInt("limit")
	if err != nil {
		return emptyStatement, err
	}

	// Offset ...
	offset, err := p.parseInt("offset")
	if err != nil {
		return emptyStatement, err
	}

	return SelectStatement{
		Table:   table,
		Items:   &items,
		Where:   where,
		GroupBy: groupBy,
		OrderBy: orderBy,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

func (p *Parser) parseSelectItems() []Expression {
	var items []Expression

	for {
		item := p.parseItem()
		if item == (Expression{}) {
			break
		}
		items = append(items, item)
		if p.matchToken(Comma) == (Token{}) {
			break
		}
	}

	return items
}

func (p *Parser) parseItem() Expression {
	for {
		var expression Expression

		item := p.matchToken(Identifier, Wildcard, Number, String)
		if item == (Token{}) {
			return expression
		}

		if item.Type == Identifier || item.Type == Wildcard {
			expression = Expression{Kind: IdentifierExpressionKind, Identifier: item.Value.(string)}
		} else {
			expression = Expression{Kind: LiteralExpressionKind, Literal: item.Value}
		}

		// If item is followed by an operator, then it's a binary expression
		operator := p.matchToken(Operator)
		if operator != (Token{}) {
			return Expression{
				Kind: BinaryExpressionKind,
				Binary: &BinaryExpression{
					A:        expression,
					B:        p.parseItem(),
					Operator: operator.Value.(string),
				},
			}
		} else {
			return expression
		}
	}
}

func (p *Parser) parseSelectTable() (string, error) {
	if !p.matchKeyword("from") {
		return "", errors.New("expected 'from' after select items")
	}
	table := p.matchToken(Identifier)
	if table == (Token{}) {
		return "", errors.New("expected identifier after 'from'")
	}
	return table.Value.(string), nil
}

func (p *Parser) parseOrderBy() (OrderBy, error) {
	var orderBy OrderBy

	if !p.matchKeyword("order by") {
		return orderBy, nil
	}
	by := p.parseItem()
	if by == (Expression{}) {
		return orderBy, errors.New("expected valid expression after 'order by'")
	}
	orderBy.By = by
	switch {
	case p.matchKeyword("desc"):
		orderBy.Direction = "desc"
	case p.matchKeyword("asc"):
		orderBy.Direction = "asc"
	default:
		orderBy.Direction = "asc"
	}
	return orderBy, nil
}

func (p *Parser) parseExpression(keywords string) (Expression, error) {
	var expression Expression

	if !p.matchKeyword(keywords) {
		return expression, nil
	}

	item := p.parseItem()
	if item == (Expression{}) {
		return expression, fmt.Errorf("expected valid expression after '%s'", keywords)
	}

	return item, nil
}

func (p *Parser) parseInt(keywords string) (int, error) {
	if !p.matchKeyword(keywords) {
		return -1, nil
	}

	item := p.parseItem()
	if item == (Expression{}) {
		return -1, fmt.Errorf("expected valid int after '%s'", keywords)
	}

	switch item.Literal.(type) {
	case int:
		return item.Literal.(int), nil
	default:
		return -1, fmt.Errorf("expected valid int after '%s", keywords)
	}
}

func (p *Parser) matchKeyword(value string) bool {
	var str string
	if p.cursor >= len(p.tokens) {
		return false
	}
	n := len(strings.Split(value, " "))
	for i := 0; i < n; i++ {
		if p.tokens[p.cursor+i].Type != Keyword {
			return false
		}
		if i != 0 {
			str += " "
		}
		str += p.tokens[p.cursor+i].Value.(string)
	}
	if str == value {
		p.cursor += n
		return true
	}
	return false
}

func (p *Parser) matchToken(tokenTypes ...TokenType) Token {
	var token Token
	if p.cursor >= len(p.tokens) {
		return token
	}
	for _, tokenType := range tokenTypes {
		if p.tokens[p.cursor].Type == tokenType {
			token = p.tokens[p.cursor]
			p.cursor++
			break
		}
	}
	return token
}
