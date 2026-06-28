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
	dbQueries      *database.Queries
	platform       string
	jwtSecret      string
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
	getJWTSecret := os.Getenv("JWT_SECRET")

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
		platform:       getPlatform,
		jwtSecret:      getJWTSecret,
	}

	httpMux := http.NewServeMux()
	fsHandler := apiCfg.middlewaresmetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))

	httpMux.Handle("/app/", fsHandler)
	httpMux.HandleFunc("GET /api/healthz", handlerReadiness)
	httpMux.HandleFunc("POST /api/users", apiCfg.handlerAddUser)
	httpMux.HandleFunc("PUT /api/users", apiCfg.handlerChangePassword)
	httpMux.HandleFunc("POST /api/login", apiCfg.handlerLogin)
	httpMux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	httpMux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	// Chirps endpoints
	httpMux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChrip)
	httpMux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	httpMux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerChripByID)
	httpMux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChrip)
	httpMux.HandleFunc("POST /api/refresh", apiCfg.handlerRefreshToken)
	httpMux.HandleFunc("POST /api/revoke", apiCfg.handlerRevokeRefreshToken)
	httpMux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerEvent)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: httpMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
