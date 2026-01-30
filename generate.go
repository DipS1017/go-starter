package main

//go:generate go tool sqlc generate .

//go:generate go tool swag init --output internal/pkg/docs --outputTypes json
//go:generate sh -c "jq '.' internal/pkg/docs/swagger.json > internal/pkg/docs/swagger.json.tmp && mv internal/pkg/docs/swagger.json.tmp internal/pkg/docs/swagger.json"
