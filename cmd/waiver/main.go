package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/stockyard-dev/stockyard-waiver/internal/server"
	"github.com/stockyard-dev/stockyard-waiver/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9801"
	}
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./waiver-data"
	}

	db, err := store.Open(dataDir)
	if err != nil {
		log.Fatalf("waiver: %v", err)
	}
	defer db.Close()

	srv := server.New(db, server.DefaultLimits())

	fmt.Printf("\n  Waiver — Self-hosted digital waiver and consent form signing\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n  Questions? hello@stockyard.dev — I read every message\n\n", port, port)
	log.Printf("waiver: listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv))
}
