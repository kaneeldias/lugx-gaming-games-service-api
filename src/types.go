package main

type GameCategory struct {
	CategoryId   int
	CategoryName string
}

type Game struct {
	GameId      int
	Name        string
	CategoryId  int
	ReleaseDate string
	Price       float64
}

type CreateGameCategoryRequest struct {
	CategoryName string
}

type CreateGameRequest struct {
	Name        string
	CategoryId  int
	ReleaseDate string
	Price       float64
}

type GetGameResponse struct {
	GameId      int     `json:"game_id"`
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	ReleaseDate string  `json:"release_date"`
	Price       float64 `json:"price"`
}

type GetGamesResponse []GetGameResponse
