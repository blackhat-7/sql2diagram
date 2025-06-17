package core

import (
	"fmt"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jmoiron/sqlx"
)

type PostgresSim struct {
	dbConnection *sqlx.DB
	pgInstance   *embeddedpostgres.EmbeddedPostgres
}

func NewPostgresSim() *PostgresSim {
	return &PostgresSim{}
}

func (p PostgresSim) connectToDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "host=localhost port=2489 user=postgres_sim password=postgres_sim dbname=postgres_sim sslmode=disable")
	return db, err
}

func (p *PostgresSim) StartSQLSim(migration string) (cleanup func(), err error) {
	if migration == "" {
		return nil, fmt.Errorf("migration string must be provided")
	}

	pgDB := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().Port(2489).Username("postgres_sim").Password("postgres_sim").Database("postgres_sim"),
	)
	if err := pgDB.Start(); err != nil {
		return nil, err
	}
	p.pgInstance = pgDB

	conn, err := p.connectToDB()
	if err != nil {
		pgDB.Stop()
		return nil, err
	}
	p.dbConnection = conn

	if _, err := conn.Exec(migration); err != nil {
		conn.Close()
		pgDB.Stop()
		return nil, err
	}

	cleanup = func() { _ = p.EndSQLSim() }
	return cleanup, nil
}

func (p *PostgresSim) ExecuteSQL(sql string) (result any, err error) {
	// For SELECT queries, we need to use Query, not Exec
	rows, err := p.dbConnection.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	// Prepare result slice
	var results [][]interface{}

	// Process each row
	for rows.Next() {
		// Create a slice to hold values for this row
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		// Create pointers to the values
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		// Convert byte arrays to strings for better handling
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				values[i] = string(b)
			}
		}

		results = append(results, values)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (p *PostgresSim) EndSQLSim() error {
	if p.dbConnection != nil {
		if err := p.dbConnection.Close(); err != nil {
			return err
		}
		p.dbConnection = nil
	}
	if p.pgInstance != nil {
		if err := p.pgInstance.Stop(); err != nil {
			return err
		}
		p.pgInstance = nil
	}
	return nil
}
