package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

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
				"message": "page loaded",
			})

		})

		api.GET("/movies", MovieHandler)
	}
	router.Run(":3000")
}

// Handlers

func MovieHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	titleList, err := loadTitles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "An internal error occurred. Please try again later."})
		return
	}
	log.Println("HANDLING")

	movieChan := make(chan Movie, len(titleList))
	errChan := make(chan error, len(titleList))
	semaphore := make(chan struct{}, 10)

	wg := sync.WaitGroup{}
	for _, title := range titleList {
		// retrieve token to run. This will block if the channel is full.
		semaphore <- struct{}{}
		go func(title string) {
			wg.Add(1)
			defer wg.Done()
			// release when we're done by removing the token from the channel
			defer func() { <-semaphore }()
			movie, err := MovieByTitle(title)
			if err != nil {
				errChan <- err
			}
			movieChan <- movie
		}(title)

	}
	wg.Wait()
	close(movieChan)
	close(errChan)
	for range cap(semaphore) {
		// wait for all token holders to complete by trying to fill
		// up the channel to its capacity
		semaphore <- struct{}{}
	}
	close(semaphore)

	var movieInfoList []Movie
	for movieInfo := range movieChan {
		movieInfoList = append(movieInfoList, movieInfo)
	}
	log.Println("Number of movies retrieved: ", len(movieInfoList))

	for err := range errChan {
		log.Println("Unable to fetch movie information by title: ", err)
	}

	c.JSON(http.StatusOK, movieInfoList)
}

func TitleHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"message": "Title not implemented!",
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
	return "http://www.omdbapi.com/?apikey=" + cfg.ApiKey + "&"
}

func MovieByTitle(title string) (Movie, error) {

	url := baseURL() + "t=" + url.QueryEscape(title)
	res, err := http.Get(url)

	if err != nil {
		log.Println("Unable to retrieve data from ", url)
		log.Println("Error:", err)
		// is returning an empty struct in an error case like this idiomatic?
		return Movie{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Unable to read response body for title: ", title)
		log.Println("Body: ", res.Body)
		log.Println("Error:", err)
		return Movie{}, err
	}
	if len(body) == 0 {
		log.Println("No additional movie info found for title: ", title)
		return Movie{Title: title}, nil
	}

	// TODO: better understand unmarshalling into different types
	var movie Movie
	var notFoundResp NotFound

	// good debugging technique! Cast body byte array to slice [:] which is then castable to string.
	// Caught an HTML response for 400 bad req!
	// log.Println("BODY: ", string(body[:]))

	err = json.Unmarshal(body, &notFoundResp)
	if err == nil && notFoundResp.Error == "Movie not found!" {
		// If we successfully unmarshall into the weird error resp omdb gives us,
		// we're in trouble
		return Movie{}, fmt.Errorf("%v was not found in OMDB!", title)
	}

	err = json.Unmarshal(body, &movie)
	if err != nil {
		log.Println("MovieByTitle: Unable to unmarshal response body for "+title+" to type Movie!", err)
		return Movie{}, err
	}

	return movie, nil
}

func loadTitles() ([]string, error) {

	file, err := os.Open("movies.txt")
	if err != nil {
		log.Print("Unable to load local list of movie titles!", err)
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var result []string
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	return result, nil

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

type NotFound struct {
	Response string `json:"Response" binding:"required"`
	Error    string `json:"Error" binding:"required"`
}
