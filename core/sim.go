package core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jmoiron/sqlx"
)

type PostgresSim struct {
	glob *string
	SqlFiles      []string
	Sql           *string
}

func NewPostgresSim() *PostgresSim {
	return &PostgresSim{}
}

func (p *PostgresSim) UseGlob(glob string) *PostgresSim {
	p.glob = &glob
	return p
}

func (p *PostgresSim) UseSqlFile(sqlFile string) *PostgresSim {
	p.SqlFiles = append(p.SqlFiles, sqlFile)
	return p
}

func (p *PostgresSim) UseSql(sql string) *PostgresSim {
	p.Sql = &sql
	return p
}

func (p *PostgresSim) Validate() error {
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
		return fmt.Errorf("only one of MigrationsDir, SqlFile, or Sql can be provided")
	} else if count == 0 {
		return fmt.Errorf("one of MigrationsDir, SqlFile, or Sql must be provided")
	}
	return nil
}

func (p PostgresSim) getSql() (string, error) {
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

func (p PostgresSim) connectToDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", "host=localhost port=2489 user=postgres_sim password=postgres_sim dbname=postgres_sim sslmode=disable")
	return db, err
}

func (p PostgresSim) Run() error {
	if err := p.Validate(); err != nil {
		return err
	}
	sql, err := p.getSql()
	if err != nil {
		return err
	}

	fmt.Printf("Making New DB\n")
	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().Port(2489).Username("postgres_sim").Password("postgres_sim").Database("postgres_sim"),
	)
	fmt.Printf("Starting DB\n")
	if err := db.Start(); err != nil {
		return err
	}
	defer func() {
		fmt.Printf("Stopping DB\n")
		if err := db.Stop(); err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Printf("Connecting to DB\n")
	conn, err := p.connectToDB()
	if err != nil {
		return err
	}

	fmt.Printf("Running migrations\n")
	if _, err := conn.Exec(sql); err != nil {
		return err
	}

	// List the tables
	fmt.Printf("Listing tables\n")
	tables, err := conn.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
	if err != nil {
		return err
	}
	fmt.Printf("Traversing tables\n")
	for tables.Next() {
		var tableName string
		err = tables.Scan(&tableName)
		if err != nil {
			return err
		}
		fmt.Printf("Table: %s\n", tableName)
	}
	fmt.Printf("Done bro\n")

	return nil
}
