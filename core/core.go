package core

import (
	"fmt"
	"os"
	"path/filepath"
)

type Schema struct {
	Tables []Table
}

type Table struct {
	Name    string
	Columns []Column
}

type Column struct {
	Name        string
	Type        string
	Constraints []Constraint
}

type Constraint interface {
	IsConstraint()
}

type DefaultContraint struct {
	Value string
}

func (d DefaultContraint) IsConstraint() {}

type PrimaryKeyContraint struct{}

func (p PrimaryKeyContraint) IsConstraint() {}

type ForeignKeyContraint struct {
	ToTable  string
	ToColumn string
}

func (f ForeignKeyContraint) IsConstraint() {}

type UniqueContraint struct{}

func (u UniqueContraint) IsConstraint() {}

type NotNullContraint struct{}

func (n NotNullContraint) IsConstraint() {}

type UnknownConstraint struct {
	Name string
}

func (u UnknownConstraint) IsConstraint() {}

type SQLSimer interface {
	StartSQLSim(migration string) (cleanup func(), err error)
	ExecuteSQL(sql string) (result any, err error)
	EndSQLSim() error
}

type SQLSchemaBuilder interface {
	BuildSchema(sqlSimer SQLSimer) Schema
}

type DiagramBuilder interface {
	BuildDiagram(schema Schema) string
	SaveDiagram(diagram string, filepath string) error
}

type MigrationBuilder struct {
	glob     *string
	SqlFiles []string
	Sql      *string
}

func NewMigrationBuilder() *MigrationBuilder {
	return &MigrationBuilder{}
}

func (p *MigrationBuilder) UseGlob(glob string) *MigrationBuilder {
	p.glob = &glob
	return p
}

func (p *MigrationBuilder) UseSqlFile(sqlFile string) *MigrationBuilder {
	p.SqlFiles = append(p.SqlFiles, sqlFile)
	return p
}

func (p *MigrationBuilder) UseSql(sql string) *MigrationBuilder {
	p.Sql = &sql
	return p
}

func (p *MigrationBuilder) Validate() error {
	// If multiple sources of sql are provided error out
	count := 0
	if p.glob != nil {
		count++
	}
	if len(p.SqlFiles) > 0 {
		count++
	}
	if p.Sql != nil {
		count++
	}

	if count > 1 {
		return fmt.Errorf("only one of glob, SqlFile, or Sql can be provided")
	} else if count == 0 {
		return fmt.Errorf("one of glob, SqlFile, or Sql must be provided")
	}
	return nil
}

func (p *MigrationBuilder) GetSql() (string, error) {
	sql := ""
	if p.Sql != nil {
		sql = *p.Sql
	} else if len(p.SqlFiles) > 0 {
		for _, sqlFile := range p.SqlFiles {
			fmt.Printf("Reading %s\n", sqlFile)
			sqlFile, err := os.ReadFile(sqlFile)
			if err != nil {
				return "", err
			}
			sql += string(sqlFile) + "\n\n"
		}
	} else if p.glob != nil {
		files, err := filepath.Glob(*p.glob)
		if err != nil {
			return "", err
		}
		for _, file := range files {
			fmt.Printf("Reading %s\n", file)
			sqlFile, err := os.ReadFile(file)
			if err != nil {
				return "", err
			}
			sql += string(sqlFile) + "\n\n"
		}
	}
	return sql, nil
}
