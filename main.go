package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	// ParseAndGenerate()
	sim := NewPostgresSim().
		UseGlob("/Users/illusion/Documents/Work/Editing/aftershoot-cloud/libs/asdb/db/migration/core/*.up.sql")
	if err := sim.Run(); err != nil {
		log.Fatal(err)
	}
}

func ParseAndGenerate() {
	// sql := "SELECT * FROM users where a = 'abc'"
	pgParser := PgParser{}
	if err := pgParser.ParseFromFile("tmp/complex2.sql"); err != nil {
		log.Fatal(err)
	}
	tables := pgParser.GetTables()
	connections := pgParser.GetConnections()
	fmt.Printf("Found %d tables and %d connections\n", len(tables), len(connections))

	// Generate D2 code for tables and save to file
	d2b := D2Builder{}
	for _, t := range tables {
		d2b.BuildTable(*t)
	}
	for _, c := range connections {
		d2b.BuildConnection(*c)
	}
	if err := os.WriteFile("tmp/complex2.d2", []byte(d2b.Generate()), 0644); err != nil {
		log.Fatal(err)
	}
}

type D2Builder struct {
	code string
}

func (d2b *D2Builder) BuildTable(table Table) {
	d2Code := ""
	d2Code += fmt.Sprintf("%s: {\n", table.TableName)
	d2Code += "\tshape: sql_table\n"
	for _, c := range table.Columns {
		if len(c.Constraints) > 0 {
			constraints := d2b.composeFieldConstraints(c.Constraints)
			d2Code += fmt.Sprintf("\t%s: %s %s\n", c.Name, c.Type, constraints)
		} else {
			d2Code += fmt.Sprintf("\t%s: %s\n", c.Name, c.Type)
		}
	}
	d2Code += "}\n"
	d2b.code += d2Code
}

func (d2b *D2Builder) BuildConnection(connection Connection) {
	d2Code := ""
	d2Code += fmt.Sprintf("%s -> %s.%s\n", connection.FromTable, connection.ToTable, connection.ToColumn)
	d2b.code += d2Code
}

func (d2b D2Builder) composeFieldConstraints(contraints []Constraint) string {
	constraints := "{ constraint: ["
	for i, c := range contraints {
		constraints += c.ToD2String()
		if i < len(contraints)-1 {
			constraints += "; "
		}
	}
	constraints += "] }"
	return constraints
}

func (d2b D2Builder) Generate() string {
	return d2b.code
}
