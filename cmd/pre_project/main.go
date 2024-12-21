package main

import "project/internal/application"

// @title Swagger Pre project API
// @version 1.16
// @description API server for information crypto coins

// @host localhost:8080
// @BasePath /

func main() {
	app, err := application.NewApp()
	if err != nil {
		panic("error create application: " + err.Error())
	}

	if err = app.Run(); err != nil {
		panic("error starting application: " + err.Error())
	}
}
