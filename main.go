package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// Router

func main() {

	router := gin.Default()

	router.Use(static.Serve("/", static.LocalFile("./views", true)))

	api := router.Group("/api")
	{

		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})

		})

		api.GET("/movies", MovieHandler)
		api.POST("/movies/like/:movieID", LikeHandler)

	}
	router.Run(":3000")
}

// Handlers

func MovieHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	title := c.Query("title")
	log.Println("Fetching movies with title", title)
	movie, err := MovieByTitle(title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "An internal error occurred. Please try again later."})
		return
	}
	c.JSON(http.StatusOK, movie)
}

func LikeHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"message": "LikeHandler not implemented!",
	})
}

// Environment Variables

type Config struct {
	ApiKey string `env:"API_KEY,required"`
}

func loadEnv() (Config, error) {
	cfg := Config{}
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	err = env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// Data Retrieval

func baseURL() string {
	cfg, err := loadEnv()
	if err != nil {
		log.Fatal("Unable to load environment variables!", err)
	}
	return "http://www.omdbapi.com/?apikey=" + cfg.ApiKey
}

func MovieByTitle(title string) (Movie, error) {
	url := baseURL() + "&t=" + title
	res, err := http.Get(url)

	if err != nil {
		log.Println("Unable to retrieve data from ", url)
		log.Println("Error:", err)
		// is returning an empty struct in an error case like this idiomatic?
		return Movie{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Unable to read response body:", res.Body)
		log.Println("Error:", err)
		return Movie{}, err
	}

	var movie Movie

	// TODO: better understand unmarshalling into different types
	err = json.Unmarshal(body, &movie)
	if err != nil {
		log.Println("Unable to unmarshal response body to type Movie!", err)
		return Movie{}, err
	}

	return movie, nil
}

// Models

type Movie struct {
	Title    string `json:"Title" binding:"required"`
	Year     string `json:"Year" binding:"required"`
	Director string `json:"Director" binding:"required"`
	// Note: Adding another field that doesn't exist in the request doesn't cause a crash.
	// Unmarshal will decode only the fields that it can find in the destination type, and fill the others with the
	// default value for that type. Checking that your fields are not the "default"
	BadField string `json:"badfield" binding:"required"`
}
