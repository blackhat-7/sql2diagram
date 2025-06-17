package main

import (
	"fmt"
	"os"
	"sql2d2/core"
)

type D2DiagramBuilder struct {
	code string
}

func NewD2DiagramBuilder() *D2DiagramBuilder {
	return &D2DiagramBuilder{}
}

// BuildDiagram implements the DiagramBuilder interface
func (d2b *D2DiagramBuilder) BuildDiagram(schema core.Schema) string {
	d2b.code = ""

	// Build tables
	for _, table := range schema.Tables {
		d2b.buildTable(table)
	}

	// Build connections (foreign key relationships)
	for _, table := range schema.Tables {
		for _, column := range table.Columns {
			for _, constraint := range column.Constraints {
				if fk, ok := constraint.(core.ForeignKeyContraint); ok {
					d2b.buildConnection(table.Name, column.Name, fk.ToTable, fk.ToColumn)
				}
			}
		}
	}

	return d2b.code
}

// SaveDiagram implements the DiagramBuilder interface
func (d2b *D2DiagramBuilder) SaveDiagram(diagram string, filepath string) error {
	return os.WriteFile(filepath, []byte(diagram), 0644)
}

func (d2b *D2DiagramBuilder) buildTable(table core.Table) {
	d2Code := ""
	d2Code += fmt.Sprintf("%s: {\n", table.Name)
	d2Code += "\tshape: sql_table\n"
	for _, column := range table.Columns {
		if len(column.Constraints) > 0 {
			constraints := d2b.composeFieldConstraints(column.Constraints)
			d2Code += fmt.Sprintf("\t%s: %s %s\n", column.Name, column.Type, constraints)
		} else {
			d2Code += fmt.Sprintf("\t%s: %s\n", column.Name, column.Type)
		}
	}
	d2Code += "}\n"
	d2b.code += d2Code
}

func (d2b *D2DiagramBuilder) buildConnection(fromTable, fromColumn, toTable, toColumn string) {
	d2Code := fmt.Sprintf("%s.%s -> %s.%s\n", fromTable, fromColumn, toTable, toColumn)
	d2b.code += d2Code
}

func (d2b *D2DiagramBuilder) composeFieldConstraints(constraints []core.Constraint) string {
	constraintStr := "{ constraint: ["
	for i, constraint := range constraints {
		constraintStr += d2b.constraintToD2String(constraint)
		if i < len(constraints)-1 {
			constraintStr += "; "
		}
	}
	constraintStr += "] }"
	return constraintStr
}

func (d2b *D2DiagramBuilder) constraintToD2String(constraint core.Constraint) string {
	switch c := constraint.(type) {
	case core.PrimaryKeyContraint:
		return "Primary Key"
	case core.ForeignKeyContraint:
		return fmt.Sprintf("Foreign Key to %s.%s", c.ToTable, c.ToColumn)
	case core.UniqueContraint:
		return "Unique"
	case core.NotNullContraint:
		return "Not Null"
	case core.DefaultContraint:
		return fmt.Sprintf("Default: %s", c.Value)
	case core.UnknownConstraint:
		return c.Name
	default:
		return "Unknown"
	}
}
