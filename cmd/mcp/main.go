package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"supplygraph/internal/db"
	"supplygraph/internal/mcpserver"
)

func main() {
	database, err := db.Open()
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	repo := db.NewRepository(database)
	server := mcpserver.New(repo).MCP()

	log.Printf("starting supplygraph mcp server on stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("run mcp server: %v", err)
	}
}
