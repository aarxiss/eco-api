package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "eco-api/docs"
	"eco-api/internal/blockchain"
	"eco-api/internal/handler"
	"eco-api/internal/repository"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}


func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)

		log.Printf("[%s] %s | Статус: %d | Час: %v | IP: %s\n",
			r.Method, r.URL.Path, lrw.statusCode, time.Since(start), r.RemoteAddr)
	})
}

// @title Measurements API
// @version 1.0
// @description API для збереження вимірювань (Чиста архітектура)
// @host localhost:8080
// @BasePath /
func main() {

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("Помилка: змінна DATABASE_URL порожня!")
	}

	dbPool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Не вдалося підключитися до БД: %v\n", err)
	}
	defer dbPool.Close()

	var bcClient *blockchain.AnchorClient
	rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
	privKey := os.Getenv("BLOCKCHAIN_PRIVATE_KEY")
	contractAddr := os.Getenv("BLOCKCHAIN_CONTRACT")

	if rpcURL != "" && privKey != "" && contractAddr != "" {
		client, err := blockchain.Init(rpcURL, privKey, contractAddr)
		if err != nil {
			log.Printf("Помилка ініціалізації блокчейну: %v\n", err)
		} else {
			bcClient = client
			log.Println("Успішно підключено до смарт-контракту в Ethereum Sepolia!")
		}
	}

	repo := repository.NewMeasurementRepo(dbPool)
	measHandler := handler.NewMeasurementHandler(repo, bcClient)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprintf(w, "Привіт з Docker! Чиста архітектура працює.")
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := dbPool.Ping(context.Background()); err != nil {
			http.Error(w, "Помилка: база даних недоступна", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Сервер живий, БД підключена!")
	})

	mux.HandleFunc("GET /measurements", measHandler.GetMeasurements)
	mux.HandleFunc("POST /measurements", measHandler.CreateMeasurement)
	mux.HandleFunc("PUT /measurements/{m_id}", measHandler.UpdateMeasurement)
	mux.HandleFunc("DELETE /measurements/{m_id}", measHandler.DeleteMeasurement)
	mux.HandleFunc("GET /measurements/{m_id}/verify", measHandler.VerifyMeasurement)
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	loggedMux := loggingMiddleware(mux)

	log.Println("Сервер працює на порту 8080...")
	log.Fatal(http.ListenAndServe(":8080", loggedMux))
}
