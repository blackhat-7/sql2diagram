package core

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
