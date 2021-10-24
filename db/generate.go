package db

//go:generate go run entgo.io/ent/cmd/entc generate --feature privacy --feature sql/upsert --target ../ent ./schema
//go:generate sh -c "go run entgo.io/ent/cmd/entc describe ./schema > schema.txt"
