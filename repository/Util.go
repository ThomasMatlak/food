package repository

import "github.com/neo4j/neo4j-go-driver/v5/neo4j/db"

func TypedGet[T any](record *db.Record, column string) (T, bool) {
	val, found := record.Get(column)

	return val.(T), found
}
