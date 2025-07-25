package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
)

func main() {
	err := InitializeDatabase()
	if err != nil {
		log.Fatal("Error initializing database: ", err)
	}

	router := gin.Default()
	router.Use(CORSMiddleware())
	router.GET("/", healthCheck)
	router.GET("/games", getAllGames)

	err = router.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
	if err != nil {
		log.Println("Error starting server: ", err)
		return
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Server is running with tag %s", os.Getenv("TAG")),
	})
}

func getAllGames(c *gin.Context) {
	games, err := GetAllGames()
	if err != nil {
		log.Println("Error fetching games: ", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, games)
}
