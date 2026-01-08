package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	db *sql.DB
)

type DorisConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func initDB(config DorisConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.Database)

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

type QueryArgs struct {
	SQL string `json:"sql" jsonschema:"SQL query to execute"`
}

type QueryOutput struct {
	Columns []string `json:"columns"`
	Rows    [][]any  `json:"rows"`
	Count   int      `json:"count"`
	Time    string   `json:"time"`
}

func executeQuery(ctx context.Context, req *mcp.CallToolRequest, args QueryArgs) (*mcp.CallToolResult, QueryOutput, error) {
	startTime := time.Now()

	if db == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Database connection not initialized"},
			},
		}, QueryOutput{}, fmt.Errorf("database not connected")
	}

	sql := strings.TrimSpace(args.SQL)
	if sql == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "SQL query cannot be empty"},
			},
		}, QueryOutput{}, fmt.Errorf("empty SQL query")
	}

	rows, err := db.QueryContext(ctx, sql)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Query failed: %v", err)},
			},
		}, QueryOutput{}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to get columns: %v", err)},
			},
		}, QueryOutput{}, err
	}

	var resultRows [][]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to scan row: %v", err)},
				},
			}, QueryOutput{}, err
		}

		row := make([]any, len(columns))
		for i, val := range values {
			if val == nil {
				row[i] = nil
			} else {
				switch v := val.(type) {
				case []byte:
					row[i] = string(v)
				case time.Time:
					row[i] = v.Format("2006-01-02 15:04:05")
				default:
					row[i] = v
				}
			}
		}
		resultRows = append(resultRows, row)
	}

	if err := rows.Err(); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error iterating rows: %v", err)},
			},
		}, QueryOutput{}, err
	}

	elapsed := time.Since(startTime)
	output := QueryOutput{
		Columns: columns,
		Rows:    resultRows,
		Count:   len(resultRows),
		Time:    elapsed.String(),
	}

	outputJSON, _ := json.MarshalIndent(output, "", "  ")
	content := fmt.Sprintf("Query executed successfully in %s\n\nColumns: %v\nRows: %d\n\n%s",
		elapsed, columns, len(resultRows), string(outputJSON))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}, output, nil
}

type ListTablesArgs struct {
	Database string `json:"database,omitempty" jsonschema:"Database name (optional, uses default if not specified)"`
}

type ListTablesOutput struct {
	Tables []string `json:"tables"`
	Count  int      `json:"count"`
}

func listTables(ctx context.Context, req *mcp.CallToolRequest, args ListTablesArgs) (*mcp.CallToolResult, ListTablesOutput, error) {
	if db == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Database connection not initialized"},
			},
		}, ListTablesOutput{}, fmt.Errorf("database not connected")
	}

	var query string
	if args.Database != "" {
		query = fmt.Sprintf("SHOW TABLES FROM `%s`", args.Database)
	} else {
		query = "SHOW TABLES"
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to list tables: %v", err)},
			},
		}, ListTablesOutput{}, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to scan table name: %v", err)},
				},
			}, ListTablesOutput{}, err
		}
		tables = append(tables, tableName)
	}

	if err := rows.Err(); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error iterating tables: %v", err)},
			},
		}, ListTablesOutput{}, err
	}

	output := ListTablesOutput{
		Tables: tables,
		Count:  len(tables),
	}

	content := fmt.Sprintf("Found %d tables:\n%s", len(tables), strings.Join(tables, "\n"))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: content},
		},
	}, output, nil
}

type DescribeTableArgs struct {
	Table    string `json:"table" jsonschema:"Table name to describe"`
	Database string `json:"database,omitempty" jsonschema:"Database name (optional)"`
}

type ColumnInfo struct {
	Field   string  `json:"field"`
	Type    string  `json:"type"`
	Null    string  `json:"null"`
	Key     string  `json:"key"`
	Default *string `json:"default,omitempty"`
	Extra   string  `json:"extra"`
}

type DescribeTableOutput struct {
	Table   string       `json:"table"`
	Columns []ColumnInfo `json:"columns"`
	Count   int          `json:"count"`
}

func describeTable(ctx context.Context, req *mcp.CallToolRequest, args DescribeTableArgs) (*mcp.CallToolResult, DescribeTableOutput, error) {
	if db == nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Database connection not initialized"},
			},
		}, DescribeTableOutput{}, fmt.Errorf("database not connected")
	}

	if args.Table == "" {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Table name is required"},
			},
		}, DescribeTableOutput{}, fmt.Errorf("table name required")
	}

	var query string
	if args.Database != "" {
		query = fmt.Sprintf("DESCRIBE `%s`.`%s`", args.Database, args.Table)
	} else {
		query = fmt.Sprintf("DESCRIBE `%s`", args.Table)
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Failed to describe table: %v", err)},
			},
		}, DescribeTableOutput{}, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var defaultVal sql.NullString
		if err := rows.Scan(&col.Field, &col.Type, &col.Null, &col.Key, &defaultVal, &col.Extra); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Failed to scan column info: %v", err)},
				},
			}, DescribeTableOutput{}, err
		}
		if defaultVal.Valid {
			col.Default = &defaultVal.String
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return &mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error iterating columns: %v", err)},
			},
		}, DescribeTableOutput{}, err
	}

	output := DescribeTableOutput{
		Table:   args.Table,
		Columns: columns,
		Count:   len(columns),
	}

	var contentBuilder strings.Builder
	contentBuilder.WriteString(fmt.Sprintf("Table: %s\nColumns: %d\n\n", args.Table, len(columns)))
	contentBuilder.WriteString("Field\tType\tNull\tKey\tDefault\tExtra\n")
	contentBuilder.WriteString(strings.Repeat("-", 80) + "\n")
	for _, col := range columns {
		defaultStr := "NULL"
		if col.Default != nil {
			defaultStr = *col.Default
		}
		contentBuilder.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n",
			col.Field, col.Type, col.Null, col.Key, defaultStr, col.Extra))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: contentBuilder.String()},
		},
	}, output, nil
}

func main() {
	host := flag.String("host", "localhost", "Doris server host")
	port := flag.Int("port", 9030, "Doris server port")
	user := flag.String("user", "root", "Doris username")
	password := flag.String("password", "", "Doris password")
	database := flag.String("database", "", "Default database name")
	flag.Parse()

	if *password == "" {
		*password = os.Getenv("DORIS_PASSWORD")
		if *password == "" {
			log.Fatal("Doris password must be provided via -password flag or DORIS_PASSWORD environment variable")
		}
	}

	if *database == "" {
		*database = os.Getenv("DORIS_DATABASE")
		if *database == "" {
			log.Fatal("Doris database must be provided via -database flag or DORIS_DATABASE environment variable")
		}
	}

	config := DorisConfig{
		Host:     *host,
		Port:     *port,
		User:     *user,
		Password: *password,
		Database: *database,
	}

	if err := initDB(config); err != nil {
		log.Fatalf("Failed to initialize database connection: %v", err)
	}
	defer db.Close()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "doris",
		Version: "v1.0.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "query",
		Description: "Execute a SQL query on Doris database and return results",
	}, executeQuery)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_tables",
		Description: "List all tables in the specified database (or default database if not specified)",
	}, listTables)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "describe_table",
		Description: "Get the structure and column information of a table",
	}, describeTable)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
