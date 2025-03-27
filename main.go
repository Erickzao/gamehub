package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Store struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Image string `json:"image"`
}

type Game struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	BackgroundImage string   `json:"background_image"`
	Genres          []string `json:"genres"`
	Rating          float64  `json:"rating"`
	RatingTop       int      `json:"rating_top"`
	ReleaseDate     string   `json:"release_date"`
	Added           int      `json:"added"`
	Metacritic      int      `json:"metacritic"`
	Playtime        int      `json:"playtime"`
	Updated         string   `json:"updated"`
	Reviews         int      `json:"reviews_count"`
	Platforms       []string `json:"platforms"`
	Stores          []Store  `json:"stores"`
}

func initializeServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Configurações de segurança
	r.Use(gin.Recovery())
	r.Use(securityHeaders())
	r.Use(limitBodySize(10 << 20)) // 10MB
	r.Use(rateLimiter())
	r.Use(validateInput())

	setupRoutes(r)
	return r
}

// securityHeaders adiciona headers de segurança para todas as respostas
func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// limitBodySize limita o tamanho do corpo das requisições
func limitBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// rateLimiter implementa limitação de taxa de requisições
func rateLimiter() gin.HandlerFunc {
	limiter := make(map[string]int64)
	var mu sync.Mutex
	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		now := time.Now().Unix()
		// Limpa entradas antigas (mais de 1 minuto)
		for key, timestamp := range limiter {
			if now-timestamp > 60 {
				delete(limiter, key)
			}
		}
		// Verifica se o IP já atingiu o limite
		if timestamp, exists := limiter[ip]; exists && now-timestamp < 1 {
			mu.Unlock()
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many requests"})
			return
		}
		limiter[ip] = now
		mu.Unlock()
		c.Next()
	}
}

// validateInput valida e sanitiza as entradas
func validateInput() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validar tamanho dos parâmetros
		if len(c.Request.URL.RawQuery) > 1024 {
			c.AbortWithStatusJSON(400, gin.H{"error": "Query string too long"})
			return
		}

		// Validar Content-Type para POSTs
		if c.Request.Method == "POST" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" {
				c.AbortWithStatusJSON(415, gin.H{"error": "Unsupported Media Type"})
				return
			}
		}

		// Validar caracteres especiais em parâmetros
		for param := range c.Request.URL.Query() {
			if strings.ContainsAny(param, "<>{}[]\\") {
				c.AbortWithStatusJSON(400, gin.H{"error": "Invalid characters in parameters"})
				return
			}
		}

		c.Next()
	}
}

func setupRoutes(r *gin.Engine) {
	r.GET("/games", getGames)
	r.GET("/games/latest", getLatestGames)
	r.GET("/games/popular", getPopularGames)
	r.GET("/games/metacritic", getMetacriticGames)
	r.GET("/games/upcoming", getUpcomingGames)
	r.GET("/games/search", searchGames)
	r.GET("/games/:id", getGameByID)
}

func main() {
	if err := loadEnv(); err != nil {
		log.Fatal(err)
	}

	r := initializeServer()
	startServer(r)
}

func loadEnv() error {
	return godotenv.Load()
}

func startServer(r *gin.Engine) {
	port := getPort()
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
		// Configurações de timeout
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		// Configurações do HTTP/2
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Esperar por sinais de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return "8080"
	}
	return port
}

func getGames(c *gin.Context) {
	games, err := fetchGames("/games")
	handleResponse(c, games, err)
}

func getLatestGames(c *gin.Context) {
	games, err := fetchGames("/games?ordering=-released")
	handleResponse(c, games, err)
}

func getPopularGames(c *gin.Context) {
	games, err := fetchGames("/games?ordering=-rating")
	handleResponse(c, games, err)
}

func getMetacriticGames(c *gin.Context) {
	games, err := fetchGames("/games?ordering=-metacritic")
	handleResponse(c, games, err)
}

func getUpcomingGames(c *gin.Context) {
	games, err := fetchGames("/games?dates=2024-03-26,2025-03-26&ordering=released")
	handleResponse(c, games, err)
}

func searchGames(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "Search query is required"})
		return
	}

	games, err := fetchGames("/games?search=" + query)
	handleResponse(c, games, err)
}

func getGameByID(c *gin.Context) {
	id := c.Param("id")
	game, err := fetchGameByID(id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch game"})
		return
	}

	if game == nil {
		c.JSON(404, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(200, game)
}

func handleResponse(c *gin.Context, data interface{}, err error) {
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(200, data)
}
