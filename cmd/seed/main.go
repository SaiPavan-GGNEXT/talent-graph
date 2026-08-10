// Command seed wipes the database and loads the demo dataset.
//
//	NEO4J_URI=bolt+s://… NEO4J_PASSWORD=… go run ./cmd/seed
package main

import (
	"context"
	"log"
	"time"

	"talentgraph/internal/config"
	"talentgraph/internal/graph"
	"talentgraph/internal/seed"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	store, err := graph.New(ctx, cfg.Neo4jURI, cfg.Neo4jUser, cfg.Neo4jPassword)
	if err != nil {
		log.Fatalf("connecting to graph database: %v", err)
	}
	defer store.Close(ctx)

	if err := seed.Load(ctx, store); err != nil {
		log.Fatalf("seed: %v", err)
	}
}
