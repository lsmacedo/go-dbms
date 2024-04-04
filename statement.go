package main

/*
---------
Statement
---------
*/

type Statement struct {
	Select      SelectStatement
	Insert      InsertStatement
	CreateTable CreateTableStatement
	Kind        StatementKind
}

type StatementKind uint

const (
	SelectKind StatementKind = iota
	InsertKind
	CreateTableKind
)

/*
----------
Expression
----------
*/

type Expression struct {
	Literal    string
	Identifier string
	Binary     *BinaryExpression
	Function   Function
	Kind       ExpressionKind
}

type ExpressionKind uint

const (
	LiteralExpressionKind ExpressionKind = iota
	IdentifierExpressionKind
	BinaryExpressionKind
	FunctionExpressionKind
)

type BinaryExpression struct {
	A        Expression
	B        Expression
	Operator string
}

type Function struct {
	Name   string
	Params *[]Expression
}

/*
----------------
Select statement
----------------
*/

type SelectStatement struct {
	Table   string
	Items   *[]Expression
	Where   Expression
	GroupBy Expression
	OrderBy OrderByExpression
	Limit   int
	Offset  int
}

type OrderByExpression struct {
	By        Expression
	Direction string
}

/*
----------------
Insert statement
----------------
*/

type InsertStatement struct {
	Table   string
	Columns *[]Expression
	Values  *[]Expression
}

/*
----------------------
Create table statement
----------------------
*/

type CreateTableStatement struct {
	Name    string
	Columns *[]ColumnDefinition
}

type ColumnDefinition struct {
	Name string
	Type string
}
