package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"sql2d2/core"

	"github.com/alecthomas/kong"
)

// CLI represents the command-line interface structure
type CLI struct {
	Input       string `arg:"" help:"SQL file or glob pattern for migration files" type:"path"`
	Output      string `short:"o" long:"output" help:"Output path for the diagram file" default:"tmp/schema.d2"`
	SqlType     string `short:"s" long:"sql-type" help:"SQL database type" enum:"postgres" default:"postgres"`
	DiagramTool string `short:"d" long:"diagram-tool" help:"Diagramming tool to use" enum:"d2" default:"d2"`
}

func main() {
	var cli CLI
	ctx := kong.Parse(&cli,
		kong.Name("sql2d2"),
		kong.Description("Generate database diagrams from SQL migration files"),
		kong.UsageOnError(),
	)

	// Build migration using the provided input
	mb := core.NewMigrationBuilder().UseGlob(cli.Input)
	sql, err := mb.GetSql()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize database simulator based on SQL type
	var sim core.SQLSimer
	switch cli.SqlType {
	case "postgres":
		sim = core.NewPostgresSim()
	default:
		ctx.FatalIfErrorf(fmt.Errorf("unsupported SQL type: %s", cli.SqlType))
	}

	cleanup, err := sim.StartSQLSim(sql)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// Build schema using appropriate schema builder
	var schemaBuilder core.SQLSchemaBuilder
	switch cli.SqlType {
	case "postgres":
		schemaBuilder = core.NewPostgresSQLSchemaBuilder()
	default:
		ctx.FatalIfErrorf(fmt.Errorf("unsupported SQL type for schema building: %s", cli.SqlType))
	}

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

	// Generate diagram using the specified tool
	var diagram string
	switch cli.DiagramTool {
	case "d2":
		diagramBuilder := NewD2DiagramBuilder()
		diagram = diagramBuilder.BuildDiagram(schema)
	default:
		ctx.FatalIfErrorf(fmt.Errorf("unsupported diagram tool: %s", cli.DiagramTool))
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(cli.Output)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Save diagram to the specified output path
	switch cli.DiagramTool {
	case "d2":
		diagramBuilder := NewD2DiagramBuilder()
		if err := diagramBuilder.SaveDiagram(diagram, cli.Output); err != nil {
			log.Fatal(err)
		}
	default:
		ctx.FatalIfErrorf(fmt.Errorf("unsupported diagram tool for saving: %s", cli.DiagramTool))
	}

	fmt.Printf("\n%s diagram saved to %s\n", cli.DiagramTool, cli.Output)
}
