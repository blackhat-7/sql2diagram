# sql2diagram

A command-line tool that generates database diagrams from SQL migration files. Convert your SQL schema definitions into beautiful, interactive diagrams using D2.

## About

sql2diagram analyzes SQL migration files and automatically generates database diagrams that visualize:

- **Tables** with their columns and data types
- **Primary Keys (PK)** and **Foreign Keys (FK)** 
- **Constraints** (UNIQUE, NOT NULL, DEFAULT values)
- **Relationships** between tables through foreign key connections

The tool uses an embedded PostgreSQL instance to parse and validate your SQL, then extracts the schema information to generate D2 diagrams that can be rendered as SVG, PNG, or viewed interactively.

**Currently Supported:**
- PostgreSQL SQL syntax
- D2 diagram generation
- Glob pattern support for multiple migration files
- Automatic relationship detection
- Constraint visualization

**Current Limitations:**
- Only PostgreSQL SQL dialect is supported
- Only D2 diagramming tool is supported

## How to Run

### Prerequisites

Ensure you have Go installed on your system.

### Installation

```bash
go build -o sql2diagram
```

### Usage

```bash
# Generate diagram from a single SQL file
./sql2diagram schema.sql

# Generate diagram from multiple migration files using glob pattern
./sql2diagram "migrations/*.up.sql"

# Specify custom output path
./sql2diagram schema.sql -o diagrams/my-schema.d2

# Specify SQL database type (currently supports postgres)
./sql2diagram schema.sql -s postgres

# Full example with all options
./sql2diagram "migrations/*.up.sql" -o output/schema.d2 -s postgres -d d2
```

### Command Line Options

- `input`: SQL file or glob pattern for migration files (required)
- `-o, --output`: Output path for the diagram file (default: `tmp/schema.d2`)
- `-s, --sql-type`: SQL database type (default: `postgres`)
- `-d, --diagram-tool`: Diagramming tool to use (default: `d2`)

### Example

```bash
# Generate diagram from all SQL files in migrations folder
./sql2diagram "migrations/*.sql" -o schema-diagram.d2
```

The tool will:
1. Read and combine all matching SQL files
2. Start an embedded PostgreSQL instance
3. Execute the SQL to create the schema
4. Extract table and constraint information
5. Generate a D2 diagram file
6. Save the diagram to the specified output path

## How It Works

sql2diagram follows a systematic approach to convert SQL into diagrams:

### 1. **SQL Processing**
- Reads SQL files (single file or glob pattern)
- Combines multiple migration files in order
- Validates SQL syntax

### 2. **Schema Simulation** 
- Starts an embedded PostgreSQL instance on port 2489
- Executes the SQL migrations to create the actual database schema
- This ensures accurate parsing of complex SQL features

### 3. **Schema Extraction**
- Queries the PostgreSQL information_schema to extract:
  - Table names and structures
  - Column names, data types, and properties
  - Primary key constraints
  - Foreign key relationships and references
  - Unique, NOT NULL, and DEFAULT constraints

### 4. **Diagram Generation**
- Maps database schema to D2 diagram syntax
- Creates visual representations of:
  - Tables as containers with column lists
  - Primary keys marked with special indicators
  - Foreign key relationships as connecting lines
  - Constraint annotations

### 5. **Output**
- Saves the generated D2 code to the specified output file
- The D2 file can then be rendered using the D2 CLI tool to create SVG, PNG, or other visual formats

## Roadmap

### SQL Database Support
- [x] **PostgreSQL** - Support for PostgreSQL database schema parsing
- [ ] **SQLite** - Support for SQLite database schema parsing
- [ ] **MySQL** - Support for MySQL database schema parsing
- [ ] **SQL Server** - Support for Microsoft SQL Server schema parsing
- [ ] **Oracle** - Support for Oracle database schema parsing

### Diagramming Tools
- [x] **D2** - Generate D2 diagrams (currently supported)
- [ ] **Mermaid** - Generate Mermaid entity relationship diagrams
- [ ] **PlantUML** - Support for PlantUML database diagrams
- [ ] **Graphviz** - Generate DOT files for Graphviz rendering

### Additional Features
- Support for views and stored procedures
- Enhanced constraint visualization
- Custom styling and theming options
- Interactive web-based diagram viewer

