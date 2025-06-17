package main

import (
	"fmt"
	"os"
	"slices"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

type Table struct {
	TableName string
	Columns   []*Column
}

type Column struct {
	Name        string
	Type        string
	Constraints []Constraint
}

type Connection struct {
	FromTable  string
	FromColumn string
	ToTable    string
	ToColumn   string
}

type Constraint interface {
	ToD2String() string
}

type DefaultContraint struct {
	Value string
}

func (d DefaultContraint) ToD2String() string {
	return fmt.Sprintf("DEFAULT:%s", d.Value)
}

type PrimaryKeyContraint struct{}

func (p PrimaryKeyContraint) ToD2String() string { return "primary_key" }

type ForeignKeyContraint struct{}

func (f ForeignKeyContraint) ToD2String() string { return "foreign_key" }

type UniqueContraint struct{}

func (u UniqueContraint) ToD2String() string { return "unique" }

type NotNullContraint struct{}

func (n NotNullContraint) ToD2String() string { return "NOTNULL" }

type UnknownConstraint struct {
	Name string
}

func (u UnknownConstraint) ToD2String() string { return u.Name }

func pgQueryToConstraint(c *pg_query.Node, columnType string) Constraint {
	constraintName := c.GetConstraint().GetContype().String()
	switch constraintName {
	case "CONSTR_DEFAULT":
		defaultValue := ""
		if columnType == "varchar" {
			defaultValue = c.GetConstraint().GetRawExpr().GetTypeCast().GetArg().GetAConst().GetSval().GetSval()
			if defaultValue == "" {
				defaultValue = "''"
			}
		} else if columnType == "bool" {
			defaultValue = c.GetConstraint().GetRawExpr().GetAConst().GetBoolval().String()
			if defaultValue == "" {
				defaultValue = "false"
			}
		} else {
			defaultValue = c.GetConstraint().GetRawExpr().GetAConst().GetIval().String()
			if defaultValue == "" {
				defaultValue = "0"
			}
		}
		return DefaultContraint{Value: defaultValue}
	case "CONSTR_PRIMARY":
		return PrimaryKeyContraint{}
	case "CONSTR_FOREIGN":
		return ForeignKeyContraint{}
	case "CONSTR_UNIQUE":
		return UniqueContraint{}
	case "CONSTR_NOTNULL":
		return NotNullContraint{}
	default:
		return UnknownConstraint{Name: constraintName}
	}
}

type PgParser struct {
	tables      []*Table
	connections []*Connection
}

func (p *PgParser) ParseFromFile(filePath string) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read sql file: %w", err)
	}

	stmt, err := pg_query.Parse(string(file))
	if err != nil {
		return fmt.Errorf("failed to parse sql file: %w", err)
	}

	for _, s := range stmt.GetStmts() {
		stmt := s.GetStmt()
		createStmt := stmt.GetCreateStmt()
		if createStmt != nil {
			err := p.parseCreateStmt(createStmt)
			if err != nil {
				return err
			}
			continue
		}
		alterTableStmt := stmt.GetAlterTableStmt()
		if alterTableStmt != nil {
			err := p.parseAlterTableStmt(alterTableStmt)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (p PgParser) GetTables() []*Table {
	return p.tables
}

func (p PgParser) GetConnections() []*Connection {
	return p.connections
}

func (p *PgParser) parseCreateStmt(createStmt *pg_query.CreateStmt) error {
	tableName := createStmt.GetRelation().GetRelname()
	fmt.Printf("Parsing create table for %s\n", tableName)
	columns := []*Column{}
	for _, c := range createStmt.GetTableElts() {
		columnDef := c.GetColumnDef()
		if columnDef == nil {
			continue
		}
		columnName := columnDef.GetColname()
		// string type
		columnType := p.pgQueryColumnType(columnDef)

		constaints := []Constraint{}
		for _, c := range columnDef.GetConstraints() {
			constaints = append(constaints, pgQueryToConstraint(c, columnType))
		}
		columns = append(columns, &Column{Name: columnName, Type: columnType, Constraints: constaints})
	}
	// Add foreign_key constraint to fields
	foreignKeys := createStmt.GetTableElts()[len(createStmt.GetTableElts())-1].GetConstraint().GetFkAttrs()
	foreignKeyColumns := []string{}
	for _, fk := range foreignKeys {
		foreignKeyColumns = append(foreignKeyColumns, fk.GetString_().GetSval())
	}
	for _, col := range columns {
		if slices.Contains(foreignKeyColumns, col.Name) {
			col.Constraints = append(col.Constraints, ForeignKeyContraint{})
		}
	}
	p.tables = append(p.tables, &Table{TableName: tableName, Columns: columns})
	for _, c := range createStmt.GetTableElts() {
		if c.GetConstraint().GetContype().String() == "CONSTR_FOREIGN" {
			p.connections = append(p.connections, &Connection{
				FromTable:  tableName,
				FromColumn: c.GetConstraint().GetFkAttrs()[0].GetString_().GetSval(),
				ToTable:    c.GetConstraint().GetPktable().GetRelname(),
				ToColumn:   c.GetConstraint().GetPkAttrs()[0].GetString_().GetSval(),
			})
		}
	}

	return nil
}

func (p *PgParser) parseAlterTableStmt(alterTableStmt *pg_query.AlterTableStmt) error {
	fmt.Printf("Parsing alter table for %s\n", alterTableStmt.GetRelation().GetRelname())
	fromTable := alterTableStmt.GetRelation().GetRelname()
	for _, c := range alterTableStmt.GetCmds() {
		alterTableCmd := c.GetAlterTableCmd()
		if alterTableCmd == nil {
			continue
		}
		subtype := alterTableCmd.GetSubtype()
		switch subtype {
		case pg_query.AlterTableType_AT_AddConstraint:
			def := alterTableCmd.GetDef()
			constraint := def.GetConstraint()
			contype := constraint.GetContype().String()
			switch contype {
			case "CONSTR_FOREIGN":
				// Add connection
				pktable := constraint.GetPktable().GetRelname()
				fkAttrs := constraint.GetFkAttrs()
				pkAttrs := constraint.GetPkAttrs()
				p.connections = append(p.connections, &Connection{
					FromTable:  fromTable,
					FromColumn: fkAttrs[0].GetString_().GetSval(),
					ToTable:    pktable,
					ToColumn:   pkAttrs[0].GetString_().GetSval(),
				})
				// Modify tables to add foreign key constraint to columns
				for _, t := range p.tables {
					for i, c := range t.Columns {
						if c.Name == fkAttrs[0].GetString_().GetSval() {
							t.Columns[i].Constraints = append(t.Columns[i].Constraints, ForeignKeyContraint{})
						}
					}
				}
			}
		}
	}
	return nil
}

func (p PgParser) pgQueryColumnType(columnDef *pg_query.ColumnDef) string {
	names := columnDef.GetTypeName().GetNames()
	var columnType string
	if names[0].GetString_().GetSval() == "pg_catalog" {
		columnType = names[1].GetString_().GetSval()
	} else {
		columnType = names[0].GetString_().GetSval()
	}
	return columnType
}
