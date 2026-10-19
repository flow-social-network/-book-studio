package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"book-studio/backend/internal/database"
	"book-studio/backend/internal/auth"
	"book-studio/backend/internal/httpapi"
)

func main() {
	addr := env("HTTP_ADDR", ":8080")
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" { log.Fatal("DATABASE_URL is required; refusing to start without a configured database") }
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil { log.Fatalf("configure PostgreSQL pool: %v", err) }
	defer pool.Close()
	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	if err := pool.Ping(pingCtx); err != nil { cancelPing(); log.Fatalf("PostgreSQL is unavailable: %v", err) }
	cancelPing()
	if err := database.Migrate(ctx, pool); err != nil { log.Fatalf("apply database migrations: %v", err) }

	firebaseClient, err := auth.NewFirebaseVerifier(ctx, os.Getenv("FIREBASE_CREDENTIALS_FILE"), os.Getenv("FIREBASE_PROJECT_ID"))
	if err != nil { log.Fatalf("initialize Firebase token verifier: %v", err) }

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status":"ok"}) })
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		pingCtx, cancel := context.WithTimeout(r.Context(), 3*time.Second); defer cancel()
		if err := pool.Ping(pingCtx); err != nil { writeJSON(w, 503, map[string]string{"status":"not_ready"}); return }
		writeJSON(w, 200, map[string]string{"status":"ready"})
	})
	protected := http.NewServeMux()
	(&httpapi.API{DB: pool}).Register(protected)
	mux.Handle("/api/v1/", auth.RequireUser(firebaseClient, protected))

	server := &http.Server{Addr: addr, Handler: cors(securityHeaders(mux), os.Getenv("CORS_ALLOWED_ORIGINS")), ReadHeaderTimeout:5*time.Second, ReadTimeout:15*time.Second, WriteTimeout:30*time.Second, IdleTimeout:60*time.Second}
	go func() { <-ctx.Done(); shutdownCtx,cancel:=context.WithTimeout(context.Background(),10*time.Second);defer cancel();if err:=server.Shutdown(shutdownCtx);err!=nil&&!errors.Is(err,http.ErrServerClosed){log.Printf("shutdown server: %v",err)} }()
	log.Printf("Book Studio API listening on %s", addr)
	if err:=server.ListenAndServe();err!=nil&&!errors.Is(err,http.ErrServerClosed){log.Fatalf("serve API: %v",err)}
}

func env(key, fallback string) string { if v:=strings.TrimSpace(os.Getenv(key));v!=""{return v};return fallback }
func writeJSON(w http.ResponseWriter,status int,payload any){w.Header().Set("Content-Type","application/json; charset=utf-8");w.WriteHeader(status);_ = json.NewEncoder(w).Encode(payload)}
func securityHeaders(next http.Handler)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){w.Header().Set("X-Content-Type-Options","nosniff");w.Header().Set("X-Frame-Options","DENY");w.Header().Set("Referrer-Policy","strict-origin-when-cross-origin");next.ServeHTTP(w,r)})}
func cors(next http.Handler, allowed string)http.Handler{return http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
	origin:=r.Header.Get("Origin")
	for _,candidate:=range strings.Split(allowed,","){candidate=strings.TrimSpace(candidate);if candidate!=""&&origin==candidate{w.Header().Set("Access-Control-Allow-Origin",origin);w.Header().Add("Vary","Origin");w.Header().Set("Access-Control-Allow-Headers","Authorization, Content-Type");w.Header().Set("Access-Control-Allow-Methods","GET, POST, PATCH, PUT, DELETE, OPTIONS");break}}
	if r.Method==http.MethodOptions { w.WriteHeader(http.StatusNoContent);return };next.ServeHTTP(w,r)
})}
