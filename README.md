# sql2diagram

A command-line tool that generates database diagrams from SQL migration files. Convert your SQL schema definitions into beautiful, interactive diagrams using D2.

## About

sql2diagram analyzes SQL migration files and automatically generates database diagrams that visualize:

- **Tables** with their columns and data types
- **Primary Keys (PK)** and **Foreign Keys (FK)** 
- **Constraints** (UNIQUE, NOT NULL, DEFAULT values)
- **Relationships** between tables through foreign key connections

The tool uses an embedded PostgreSQL instance to parse and validate your SQL, then extracts the schema information to generate D2 diagrams that can be rendered as SVG, PNG, or viewed interactively.

**Supported Features:**
- PostgreSQL SQL syntax
- D2 diagram generation
- Glob pattern support for multiple migration files
- Automatic relationship detection
- Constraint visualization

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
./sql2diagram "migrations/*.sql"

# Specify custom output path
./sql2diagram schema.sql -o diagrams/my-schema.d2

# Specify SQL database type (currently supports postgres)
./sql2diagram schema.sql -s postgres

# Full example with all options
./sql2diagram "migrations/*.sql" -o output/schema.d2 -s postgres -d d2
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

### Architecture

The tool is built with a modular architecture:

- **Core Package**: Defines interfaces and data structures
- **SQL Simulator**: Handles database instance management (PostgreSQL)
- **Schema Builder**: Extracts schema information from the database
- **Diagram Builder**: Converts schema to D2 format
- **Migration Builder**: Handles file reading and SQL preparation

This design allows for easy extension to support additional database types and diagram formats in the future.

