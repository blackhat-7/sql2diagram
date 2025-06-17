package main

import (
	"fmt"
	"log"
	"os"

	"sql2d2/core"
)

func main() {
	// ParseAndGenerate()

	mb := core.NewMigrationBuilder().
		UseGlob("/Users/illusion/Documents/Work/Editing/aftershoot-cloud/libs/asdb/db/migration/core/*.up.sql")
	sql, err := mb.GetSql()
	if err != nil {
		log.Fatal(err)
	}

	sim := core.NewPostgresSim()
	cleanup, err := sim.StartSQLSim(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// Build schema using PostgresSQLSchemaBuilder
	schemaBuilder := core.NewPostgresSQLSchemaBuilder()
	schema := schemaBuilder.BuildSchema(sim)

	// Print the schema information
	fmt.Printf("Found %d tables:\n", len(schema.Tables))
	for _, table := range schema.Tables {
		fmt.Printf("\nTable: %s\n", table.Name)
		for _, column := range table.Columns {
			fmt.Printf("  Column: %s (%s)", column.Name, column.Type)
			if len(column.Constraints) > 0 {
				fmt.Printf(" - Constraints: ")
				for i, constraint := range column.Constraints {
					if i > 0 {
						fmt.Printf(", ")
					}
					switch c := constraint.(type) {
					case core.PrimaryKeyContraint:
						fmt.Printf("PK")
					case core.ForeignKeyContraint:
						fmt.Printf("FK->%s.%s", c.ToTable, c.ToColumn)
					case core.UniqueContraint:
						fmt.Printf("UNIQUE")
					case core.NotNullContraint:
						fmt.Printf("NOT NULL")
					case core.DefaultContraint:
						fmt.Printf("DEFAULT(%s)", c.Value)
					case core.UnknownConstraint:
						fmt.Printf("%s", c.Name)
					}
				}
			}
			fmt.Printf("\n")
		}
	}

	// Generate D2 diagram and save to tmp
	diagramBuilder := NewD2DiagramBuilder()
	diagram := diagramBuilder.BuildDiagram(schema)

	// Ensure tmp directory exists
	if err := os.MkdirAll("tmp", 0755); err != nil {
		log.Fatal(err)
	}

	// Save diagram to tmp directory
	if err := diagramBuilder.SaveDiagram(diagram, "tmp/schema.d2"); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nD2 diagram saved to tmp/schema.d2\n")
}

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

// Legacy parsing function using the parse.go types
// func ParseAndGenerate() {
// 	// sql := "SELECT * FROM users where a = 'abc'"
// 	pgParser := PgParser{}
// 	if err := pgParser.ParseFromFile("tmp/complex2.sql"); err != nil {
// 		log.Fatal(err)
// 	}
// 	tables := pgParser.GetTables()
// 	connections := pgParser.GetConnections()
// 	fmt.Printf("Found %d tables and %d connections\n", len(tables), len(connections))

// 	// Generate D2 code for tables and save to file
// 	d2b := D2Builder{}
// 	for _, t := range tables {
// 		d2b.BuildTable(*t)
// 	}
// 	for _, c := range connections {
// 		d2b.BuildConnection(*c)
// 	}
// 	if err := os.WriteFile("tmp/complex2.d2", []byte(d2b.Generate()), 0644); err != nil {
// 		log.Fatal(err)
// 	}
// }

// type D2Builder struct {
// 	code string
// }

// func (d2b *D2Builder) BuildTable(table Table) {
// 	d2Code := ""
// 	d2Code += fmt.Sprintf("%s: {\n", table.TableName)
// 	d2Code += "\tshape: sql_table\n"
// 	for _, c := range table.Columns {
// 		if len(c.Constraints) > 0 {
// 			constraints := d2b.composeFieldConstraints(c.Constraints)
// 			d2Code += fmt.Sprintf("\t%s: %s %s\n", c.Name, c.Type, constraints)
// 		} else {
// 			d2Code += fmt.Sprintf("\t%s: %s\n", c.Name, c.Type)
// 		}
// 	}
// 	d2Code += "}\n"
// 	d2b.code += d2Code
// }

// func (d2b *D2Builder) BuildConnection(connection Connection) {
// 	d2Code := ""
// 	d2Code += fmt.Sprintf("%s -> %s.%s\n", connection.FromTable, connection.ToTable, connection.ToColumn)
// 	d2b.code += d2Code
// }

// func (d2b D2Builder) composeFieldConstraints(contraints []Constraint) string {
// 	constraints := "{ constraint: ["
// 	for i, c := range contraints {
// 		constraints += c.ToD2String()
// 		if i < len(contraints)-1 {
// 			constraints += "; "
// 		}
// 	}
// 	constraints += "] }"
// 	return constraints
// }

// func (d2b D2Builder) Generate() string {
// 	return d2b.code
// }
