package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Chinchzilla/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const filepathRoot = "."
const port = "8080"

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQeries       *database.Queries
	platform       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Connection to the DB failed: %s", err)
		os.Exit(1)
	}

	getPlatform := os.Getenv("PLATFORM")

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQeries:       database.New(db),
		platform:       getPlatform,
	}

	httpMux := http.NewServeMux()
	fsHandler := apiCfg.middlewaresmetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))

	httpMux.Handle("/app/", fsHandler)
	httpMux.HandleFunc("GET /api/healthz", handlerReadiness)
	httpMux.HandleFunc("POST /api/users", apiCfg.handlerAddUser)
	httpMux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	httpMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	httpMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	// Chirps endpoints
	httpMux.HandleFunc("POST /api/chirps", apiCfg.handlerNewChrip)
	httpMux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	httpMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChripByID)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: httpMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
