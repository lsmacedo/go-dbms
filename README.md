# go-dbms

A simple relational database management system (RDBMS) written in Go.

<img width="688" alt="image" src="https://github.com/lsmacedo/go-dbms/assets/29143487/10edb2fd-88b2-4d6b-9d90-e35157497a66">

## Features in scope

- [x] Column types: `integer` and `text`
- [x] Commands: `create table`, `insert` and `select`
- [x] Select clauses: `where`, `group by`, `order by`, `limit` and `offset`
- [ ] Store data on disk, organizing it into pages
- [ ] Aggregate functions
- [ ] Indexes
- [ ] Query planner
- [ ] Alias
- [ ] Joins
- [ ] Update and delete commands
- [ ] Subqueries

## Syntax

The syntax used here is a simplified version of what's usually seen in a RDBMS.

### Create table

```sql
create table `TableName` [ `ColumnName` `DataType` [, ...] ]
```

### Insert

```sql
insert into `TableName` [ `ColumnName` [, ...] ] values [ `Expression` [, ...] ]
```

### Select

```sql
select
  [ * | `Expression` [, ...] ]
  from `TableName`
  [ where `Expression` ]
  [ group by `Expression` ]
  [ order by `Expression` [ ASC | DESC ] ]
  [ limit `Limit` ]
  [ offset `Offset` ]
```

## How it works

First step is lexing and parsing the input string into a statement.
The statement is then passed into the `Backend.Run` method, which executes the
data definition, manipulation or querying.

The backend holds a map of tables, where the key is the table name. Each table
contains its column definitions and all inserted data as a bytes array.

### The data array

Data is a bytes array with similar format to how it will be stored on disk.

All rows are stored sequentially within this array, and each row is prefixed by
a 4 bytes integer indicating the offset to the next row.

Values of type `text` have a variable length. The value is prefixed by a 2 bytes
integer indicating the text size.
