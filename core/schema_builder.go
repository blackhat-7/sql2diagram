package core

import (
	"fmt"
	"strings"
)

type PostgresSQLSchemaBuilder struct{}

func NewPostgresSQLSchemaBuilder() *PostgresSQLSchemaBuilder {
	return &PostgresSQLSchemaBuilder{}
}

func (p *PostgresSQLSchemaBuilder) BuildSchema(sqlSimer SQLSimer) Schema {
	sql := `
SELECT 
    c.table_schema,
    c.table_name,
    c.column_name,
    c.data_type,
    c.is_nullable,
    c.column_default,
    ARRAY_AGG(
        CASE 
            WHEN tc.constraint_type = 'PRIMARY KEY' THEN 'PK'
            WHEN tc.constraint_type = 'FOREIGN KEY' THEN 
                'FK->' || ccu.table_name || '.' || ccu.column_name
            WHEN tc.constraint_type = 'UNIQUE' THEN 'UNIQUE'
            WHEN tc.constraint_type = 'CHECK' THEN 'CHECK'
        END
    ) FILTER (WHERE tc.constraint_type IS NOT NULL) AS constraints
FROM information_schema.columns c
LEFT JOIN information_schema.key_column_usage kcu 
    ON c.table_schema = kcu.table_schema 
    AND c.table_name = kcu.table_name 
    AND c.column_name = kcu.column_name
LEFT JOIN information_schema.table_constraints tc 
    ON kcu.constraint_name = tc.constraint_name 
    AND kcu.table_schema = tc.table_schema
LEFT JOIN information_schema.constraint_column_usage ccu 
    ON tc.constraint_name = ccu.constraint_name 
    AND tc.table_schema = ccu.table_schema
    AND tc.constraint_type = 'FOREIGN KEY'
WHERE c.table_schema = 'public'
GROUP BY c.table_schema, c.table_name, c.column_name, c.data_type, 
         c.is_nullable, c.column_default, c.ordinal_position
ORDER BY c.table_schema, c.table_name, c.ordinal_position;`

	result, err := sqlSimer.ExecuteSQL(sql)
	if err != nil {
		fmt.Printf("Error executing SQL: %v\n", err)
		return Schema{}
	}

	return p.parseQueryResult(result)
}

func (p *PostgresSQLSchemaBuilder) parseQueryResult(result any) Schema {
	// Assuming result is a slice of rows, where each row is a slice of values
	rows, ok := result.([][]any)
	if !ok {
		fmt.Printf("Unexpected result format\n")
		return Schema{}
	}

	tableMap := make(map[string]*Table)
	schema := Schema{}

	for _, row := range rows {
		if len(row) < 7 {
			continue
		}

		tableName := fmt.Sprintf("%v", row[1])
		columnName := fmt.Sprintf("%v", row[2])
		dataType := fmt.Sprintf("%v", row[3])
		isNullable := fmt.Sprintf("%v", row[4])
		columnDefault := row[5]
		constraintsRaw := row[6]

		// Get or create table
		table, exists := tableMap[tableName]
		if !exists {
			table = &Table{
				Name:    tableName,
				Columns: []Column{},
			}
			tableMap[tableName] = table
		}

		// Parse constraints
		constraints := p.parseConstraints(constraintsRaw, isNullable, columnDefault)

		// Create column
		column := Column{
			Name:        columnName,
			Type:        dataType,
			Constraints: constraints,
		}

		table.Columns = append(table.Columns, column)
	}

	// Convert map to slice
	for _, table := range tableMap {
		schema.Tables = append(schema.Tables, *table)
	}

	return schema
}

func (p *PostgresSQLSchemaBuilder) parseConstraints(constraintsRaw any, isNullable string, columnDefault any) []Constraint {
	var constraints []Constraint

	// Handle NOT NULL constraint
	if isNullable == "NO" {
		constraints = append(constraints, NotNullContraint{})
	}

	// Handle DEFAULT constraint
	if columnDefault != nil && fmt.Sprintf("%v", columnDefault) != "" {
		constraints = append(constraints, DefaultContraint{
			Value: fmt.Sprintf("%v", columnDefault),
		})
	}

	// Parse constraint array from PostgreSQL
	if constraintsRaw == nil {
		return constraints
	}

	constraintStr := fmt.Sprintf("%v", constraintsRaw)
	if constraintStr == "" || constraintStr == "{NULL}" {
		return constraints
	}

	// Remove braces and split by comma
	constraintStr = strings.Trim(constraintStr, "{}")
	if constraintStr == "" {
		return constraints
	}

	constraintList := strings.Split(constraintStr, ",")
	for _, constraint := range constraintList {
		constraint = strings.TrimSpace(constraint)
		if constraint == "" || constraint == "NULL" {
			continue
		}

		switch {
		case constraint == "PK":
			constraints = append(constraints, PrimaryKeyContraint{})
		case constraint == "UNIQUE":
			constraints = append(constraints, UniqueContraint{})
		case strings.HasPrefix(constraint, "FK->"):
			// Parse FK->table.column format
			fkPart := strings.TrimPrefix(constraint, "FK->")
			parts := strings.Split(fkPart, ".")
			if len(parts) == 2 {
				constraints = append(constraints, ForeignKeyContraint{
					ToTable:  parts[0],
					ToColumn: parts[1],
				})
			}
		default:
			constraints = append(constraints, UnknownConstraint{
				Name: constraint,
			})
		}
	}

	return constraints
}
