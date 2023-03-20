package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/db"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	"github.com/rs/zerolog/log"
)

// TODO accept context and driver
func RunQuery[T any](action string, work func(*string, *map[string]any) (T, error)) (T, error) {
	// TODO create session
	var query string
	var params map[string]any

	// TODO pass context and session to work()
	result, err := work(&query, &params)

	log.Debug().
		Str("action", action).
		Str("query", query).
		Any("params", params).
		Any("result", result).
		Err(err).
		Msg("run query")
	return result, err
}

func TypedGet[T any](record *db.Record, column string) (T, bool) {
	val, found := record.Get(column)
	return val.(T), found
}

func MatchNodeById(name string, labels []string) string {
	return fmt.Sprintf("MATCH (`%s`:`%s` {id: $%sId})", name, strings.Join(labels, "`:`"), name)
}

func RunAndReturnSingleRecord(ctx context.Context, tx neo4j.ManagedTransaction, query string, params map[string]any) (*db.Record, error) {
	result, err := tx.Run(ctx, query, params)
	if err != nil {
		return nil, err
	}

	record, err := result.Single(ctx)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func ParseResourceEntity(node dbtype.Entity) (*model.Resource, error) {
	rawCreated, err := neo4j.GetProperty[neo4j.LocalDateTime](node, "created")
	if err != nil {
		return nil, err
	}
	created := new(time.Time)
	*created = rawCreated.Time()

	rawLastModified, err := neo4j.GetProperty[neo4j.LocalDateTime](node, "lastModified")
	lastModified := new(time.Time)
	if err != nil {
		lastModified = nil
	} else {
		*lastModified = rawLastModified.Time()
	}

	return &model.Resource{Created: created, LastModified: lastModified}, nil
}
