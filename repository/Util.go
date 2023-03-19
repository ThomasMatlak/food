package repository

import (
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/db"
)

func TypedGet[T any](record *db.Record, column string) (T, bool) {
	val, found := record.Get(column)
	return val.(T), found
}

func MatchLabelById(name string, labels []string) string {
	return fmt.Sprintf("MATCH (`%s`:`%s` {id: $id})", name, strings.Join(labels, "`:`"))
}
