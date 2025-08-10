package main

import "github.com/vladazn/danish/app"

//go:generate go run github.com/swaggo/swag/cmd/swag@latest init -o ./docs

func main() {
	app.Run()
}
