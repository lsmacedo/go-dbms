# go-dbms

A simple relational database management system (RDBMS) written in Go.

<img width="688" alt="image" src="https://github.com/lsmacedo/go-dbms/assets/29143487/10edb2fd-88b2-4d6b-9d90-e35157497a66">

## Features in scope

- [x] Column types: `integer` and `text`
- [x] Commands: `create table`, `insert` and `select`
- [x] Select clauses: `where`, `group by`, `order by`, `limit` and `offset`
- [x] Store data on disk
- [ ] Cache recently accessed pages
- [ ] Aggregate functions
- [ ] Add tests
- [ ] Indexes
- [ ] Query planner
- [ ] Alias
- [ ] Joins
- [ ] Update and delete commands
- [ ] Subqueries
- [ ] Locking

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
The statement is then passed into the `Backend.Run` method, which will
execute the commands.

### Steps for creating a table:

1.  Add into table definitions the table name and its columns

### Steps for inserting data:

1.  Find the table's latest page
2.  If there is no page, or if it the row doesn't fit on it, create a new page
3.  Append data into page

### Steps for querying data:

1.  Get index of pages to select from, and iterate through them:
    1. Load page into memory
    2. Iterate through page rows:
       1. Apply filters
       2. Apply group by and group functions
       3. Evaluate and select specified items from statement
2.  Sort results
3.  Apply limit and offset

## Data

All data is currently stored on a single file called `data`, with the following
structure:

- Table definitions (table name + columns)
- Pages list (table name + cursor)
- Data pages

Rows are stored sequentially inside pages, and their values are sorted in the order
that the columns are defined. Since values may have variable length, rows have an
offset prefix.
