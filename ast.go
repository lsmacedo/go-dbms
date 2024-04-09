package main

type ExpressionKind uint

const (
	LiteralExpressionKind ExpressionKind = iota
	IdentifierExpressionKind
	BinaryExpressionKind
	FunctionExpressionKind
)

type Expression struct {
	Literal    interface{}
	Identifier string
	Binary     *BinaryExpression
	Function   Function
	Kind       ExpressionKind
}

type BinaryExpression struct {
	A        Expression
	B        Expression
	Operator string
}

type Function struct {
	Name   string
	Params *[]Expression
}

type StatementKind uint

const (
	SelectKind StatementKind = iota
	InsertKind
	CreateTableKind
)

type Statement struct {
	Select      SelectStatement
	Insert      InsertStatement
	CreateTable CreateTableStatement
	Kind        StatementKind
}

type SelectStatement struct {
	Table   string
	Items   *[]Expression
	Where   Expression
	GroupBy Expression
	OrderBy OrderBy
	Limit   int
	Offset  int
}

type OrderBy struct {
	By        Expression
	Direction string
}

type InsertStatement struct {
	Table   string
	Columns *[]Expression
	Values  *[]Expression
}

type CreateTableStatement struct {
	Name    string
	Columns *[]ColumnDefinition
}

type ColumnDefinition struct {
	Name string
	Type string
}
