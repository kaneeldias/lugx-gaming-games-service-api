package main

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"log"
	"os"
	"sync"
)

var (
	db   *sql.DB
	once sync.Once
)

func createConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_DB"))

	log.Println("Connecting to PostgreSQL...")

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	log.Println("Successfully connected to the database")
	return db, nil
}

func GetDatabaseConnection() (*sql.DB, error) {
	var err error
	once.Do(func() {
		db, err = createConnection()
	})

	if err != nil {
		return nil, fmt.Errorf("error creating database connection: %w", err)
	}

	if db == nil {
		return nil, fmt.Errorf("failed to create or retrieve database connection")
	}

	return db, err
}

func InitializeDatabase() error {
	db, err := GetDatabaseConnection()
	if err != nil {
		return fmt.Errorf("error creating database connection: %w", err)
	}

	var row string
	err = db.QueryRow(`
		SELECT name
		FROM Games
		WHERE game_id = $1
	`, 1).Scan(&row)
	if err == nil {
		log.Println("Games table already exists. Skipping initialization.")
		return nil
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS GameCategories (
		    category_id SERIAL PRIMARY KEY,
		    name VARCHAR(255) NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating categories table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS Games (
			game_id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
		    category_id INTEGER REFERENCES GameCategories(category_id),
		    release_date DATE NOT NULL,
		    price DECIMAL(5,2) NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("error creating games table: %w", err)
	}

	exploration, err := CreateGameCategory(CreateGameCategoryRequest{
		CategoryName: "Exploration",
	})
	if err != nil {
		return fmt.Errorf("error creating exploration game category: %w", err)
	}

	shooter, err := CreateGameCategory(CreateGameCategoryRequest{
		CategoryName: "Shooter",
	})
	if err != nil {
		return fmt.Errorf("error creating exploration game category: %w", err)
	}

	_, err = CreateGame(CreateGameRequest{
		Name:        "Minecraft",
		CategoryId:  exploration.CategoryId,
		ReleaseDate: "2011-01-01",
		Price:       26.95,
	})
	if err != nil {
		return fmt.Errorf("error creating Minecraft game: %w", err)
	}

	_, err = CreateGame(CreateGameRequest{
		Name:        "Counter Strike",
		CategoryId:  shooter.CategoryId,
		ReleaseDate: "1999-01-01",
		Price:       14.99,
	})

	log.Println("Database initialized.")
	return nil
}

func CreateGameCategory(request CreateGameCategoryRequest) (GameCategory, error) {
	db, err := GetDatabaseConnection()
	if err != nil {
		return GameCategory{}, fmt.Errorf("error creating database connection: %w", err)
	}

	var id int
	err = db.QueryRow(`
		INSERT INTO GameCategories (name)
		VALUES ($1)
		RETURNING category_id
	`, request.CategoryName).Scan(&id)
	if err != nil {
		return GameCategory{}, fmt.Errorf("error inserting game category: %w", err)
	}

	category := GameCategory{
		CategoryId:   id,
		CategoryName: request.CategoryName,
	}

	log.Printf("Game category created: %+v\n", category)
	return category, nil
}

func CreateGame(request CreateGameRequest) (Game, error) {
	db, err := GetDatabaseConnection()
	if err != nil {
		return Game{}, fmt.Errorf("error creating database connection: %w", err)
	}

	var id int
	err = db.QueryRow(`
		INSERT INTO Games (name, category_id, release_date, price)
		VALUES ($1, $2, $3, $4)
		RETURNING game_id
	`, request.Name, request.CategoryId, request.ReleaseDate, request.Price).Scan(&id)
	if err != nil {
		return Game{}, fmt.Errorf("error inserting game: %w", err)
	}

	game := Game{
		GameId:      id,
		Name:        request.Name,
		CategoryId:  request.CategoryId,
		ReleaseDate: request.ReleaseDate,
		Price:       request.Price,
	}

	log.Printf("Game created: %+v\n", game)
	return game, nil
}

func GetAllGames() (GetGamesResponse, error) {
	db, err := GetDatabaseConnection()
	if err != nil {
		return nil, fmt.Errorf("error creating database connection: %w", err)
	}

	rows, err := db.Query(`
		SELECT g.game_id, g.name, gc.name AS category, g.release_date, g.price
		FROM Games g
		JOIN GameCategories gc ON g.category_id = gc.category_id
	`)
	if err != nil {
		return nil, fmt.Errorf("error querying games: %w", err)
	}
	defer rows.Close()

	var games GetGamesResponse
	for rows.Next() {
		var game GetGameResponse
		err = rows.Scan(&game.GameId, &game.Name, &game.Category, &game.ReleaseDate, &game.Price)
		if err != nil {
			return nil, fmt.Errorf("error scanning game row: %w", err)
		}
		games = append(games, game)
	}

	return games, nil
}
