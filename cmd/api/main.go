package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"reporadar/internal/ai"
	"reporadar/internal/api"
	"reporadar/internal/db"
	"reporadar/internal/ollama"
	"reporadar/internal/scanjobs"
)

func main() {
	database, err := db.Open()
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	repo := db.NewRepository(database)
	runner := scanjobs.NewRunner(repo)
	if err := runner.ResumeQueuedJobs(context.Background()); err != nil {
		log.Fatalf("resume queued scan jobs: %v", err)
	}

	ollamaClient := ollama.NewClient()
	chatAI := ai.NewService(repo, ollamaClient)
	server := api.NewServer(repo, runner, chatAI)

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("starting api server on %s", addr)
	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("serve http: %v", err)
	}
}
