package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/repository"
	"github.com/ThomasMatlak/food/util"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFoodRepository(t *testing.T) {
	ctx := context.Background()

	neo4jContainer, err := startNeo4j(ctx, t)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	neo4jDriver, err := neo4jDriver(ctx, t, neo4jContainer)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	repo := repository.NewFoodRepository(*neo4jDriver)

	t.Run("Get One", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetOneFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Get One (does not exist)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetOneDoesNotExistFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Get All", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetAllFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Get All (empty)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testGetAllEmptyFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Create", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testCreateFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Update", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testUpdateFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Delete (no connections)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testDeleteNoConnectionsFood(ctx, neo4jDriver, repo, t)
	})
	t.Run("Delete (with connections)", func(t *testing.T) {
		t.Cleanup(func() { clearNeo4j(ctx, neo4jDriver) })
		testDeleteWithConnectionsFood(ctx, neo4jDriver, repo, t)
	})
}

func testGetOneFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// seed data
	id := "123"

	query := "CREATE (:Food {id: $id, name: $name, created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"id":      id,
		"name":    "test food",
		"created": neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	food, found, err := repo.GetById(ctx, id)

	assert := assert.New(t)
	assert.NoError(err)
	assert.True(found)
	assert.Equal(id, food.Id)
	assert.Equal("test food", food.Name)
	assert.WithinDuration(createdTime, *food.Created, 0)
	assert.Nil(food.LastModified)
	assert.Nil(food.Deleted)
}

func testGetOneDoesNotExistFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// no seed data

	// test
	food, found, err := repo.GetById(ctx, "test id")

	assert := assert.New(t)
	assert.NoError(err)
	assert.False(found)
	assert.Nil(food)
}

func testGetAllFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// seed data
	seedFoods := []model.Food{
		{Id: "123", Name: "test food 1"},
		{Id: "456", Name: "test food 2"},
	}

	input := util.MapArray(seedFoods, func(i model.Food) map[string]string {
		return map[string]string{"id": i.Id, "name": i.Name}
	})

	query := "UNWIND $input AS i CREATE (:Food {id: i.id, name: i.name, created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"input":   input,
		"created": neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	foods, err := repo.GetAll(ctx)

	assert := assert.New(t)
	assert.NoError(err)
	// comparing time stamps is tricky
	foodsWithoutCreated := util.MapArray(foods, func(i model.Food) model.Food {
		return model.Food{Id: i.Id, Name: i.Name}
	})
	assert.ElementsMatch(seedFoods, foodsWithoutCreated)
}

func testGetAllEmptyFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// no seed data

	// test
	foods, err := repo.GetAll(ctx)

	assert := assert.New(t)
	assert.NoError(err)
	assert.Empty(foods)
}

func testCreateFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	name := "test food"
	food := model.Food{Name: name}
	createdFood, err := repo.Create(ctx, food)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.NotEmpty(createdFood.Id)
	assert.Equal(name, createdFood.Name)
	assert.WithinDuration(time.Now(), *createdFood.Created, time.Duration(1_000_000_000))
	assert.Nil(createdFood.LastModified)
	assert.Nil(createdFood.Deleted)
}

func testUpdateFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// seed data
	id := "123"

	query := "CREATE (:Food {id: $id, name: $name, created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"id":      id,
		"name":    "test food",
		"created": neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	food := model.Food{Id: id, Name: "test food updated"}
	updatedFood, err := repo.Update(ctx, food)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, updatedFood.Id)
	assert.Equal("test food updated", updatedFood.Name)
	assert.WithinDuration(createdTime, *updatedFood.Created, 0)
	assert.True((*updatedFood.LastModified).After(createdTime))
	assert.Nil(updatedFood.Deleted)
}

func testDeleteNoConnectionsFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// seed data
	id := "123"

	query := "CREATE (:Food {id: $id, name: $name, created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"id":      id,
		"name":    "test food",
		"created": neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	deletedId, err := repo.Delete(ctx, id)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, deletedId)
}

func testDeleteWithConnectionsFood(ctx context.Context, neo4jDriver *neo4j.DriverWithContext, repo model.FoodRepository, t *testing.T) {
	// seed data
	id := "123"

	// TODO an undeleted :CONTAINS_RELATIONSHIP pointing at the food to delete should cause an error
	query := "CREATE (:Food {id: $id, name: $name, created: $created})<-[:CONTAINS_INGREDIENT {created: $created}]-(:Recipe:Resource {created: $created})"
	createdTime := time.Now()
	params := map[string]any{
		"id":      id,
		"name":    "test food",
		"created": neo4j.LocalDateTime(createdTime),
	}

	_, err := neo4j.ExecuteWrite(ctx, (*neo4jDriver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, query, params)
		})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// test
	deletedId, err := repo.Delete(ctx, id)

	// TODO make a direct Cypher query to verify the relationship was marked as deleted OR store an id on relationships and return the deleted relationship ids

	assert := assert.New(t)
	assert.NoError(err)
	assert.Equal(id, deletedId)
}

func clearNeo4j(ctx context.Context, driver *neo4j.DriverWithContext) (neo4j.ResultWithContext, error) {
	return neo4j.ExecuteWrite(ctx, (*driver).NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}),
		func(tx neo4j.ManagedTransaction) (neo4j.ResultWithContext, error) {
			return tx.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
		})
}

func neo4jDriver(ctx context.Context, t *testing.T, container *testcontainers.Container) (*neo4j.DriverWithContext, error) {
	ip, err := (*container).Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := (*container).MappedPort(ctx, "7687")
	if err != nil {
		return nil, err
	}

	driver, err := neo4j.NewDriverWithContext(fmt.Sprintf("bolt://%s:%s", ip, port.Port()), neo4j.NoAuth())
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		if err := driver.Close(ctx); err != nil {
			t.Fatalf("failed to close neo4j driver: %s", err)
		}
	})

	return &driver, nil
}

func startNeo4j(ctx context.Context, t *testing.T) (*testcontainers.Container, error) {
	return startContainer(ctx, t,
		"neo4j:5.5.0-community",
		[]string{"7687/tcp", "7474/tcp"},
		map[string]string{"NEO4J_AUTH": "none"},
		wait.ForAll(
			wait.ForLog("Started."),
			wait.ForListeningPort("7687/tcp"),
			wait.ForListeningPort("7474/tcp"),
		))
}

// TODO use docker compose https://golang.testcontainers.org/features/docker_compose/
func startContainer(
	ctx context.Context,
	t *testing.T,
	image string,
	ports []string,
	env map[string]string,
	waitStrategy wait.Strategy,
) (*testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: ports,
		Env:          env,
		WaitingFor:   waitStrategy,
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate %s: %s", image, err)
		}
	})

	return &container, nil
}
