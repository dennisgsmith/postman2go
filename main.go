//go:build ignore

package main

import (
	"postman2go/postman"
)

func main() {
	variables := make(map[string]string)
	// add postman env variables to map
	// e.g.
	// variables["base_url"] = "http://localhost:8080"
	// ...

	// create the postman config
	configs := postman.Config{
		Package:     "server",
		PostmanFile: "./path/to/postmant-collection.json",
		TestFile:    "api_test.go",
		// Add code to setup test router
		// e.g.
		// SetupRouter: `
		//	cfg := config.NewApiConfig()
		//	h := handler.NewApi(cfg)
		//	s := NewApiServer(cfg, h)
		// `,
		SetupRouter: ``,
		// RouterFunc (field of ApiServer)
		// e.g. (using api server variable "s" above
		// RouterFunc: "s.e",
		RouterFunc:        "",
		Variables:         variables,
		AdditionalImports: ``,
	}

	// generate the Golang Postman Tests
	err := configs.Generate()
	if err != nil {
		panic(err)
	}
}
