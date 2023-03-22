package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ThomasMatlak/food/model"
	"github.com/ThomasMatlak/food/repository"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TODO use subtests https://go.dev/blog/subtests if possible to avoid needing to spin up a ton of containers

func TestIngredientCreate(t *testing.T) {
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

	repo := repository.NewIngredientRepository(*neo4jDriver)

	ingredient := model.Ingredient{Name: "test ingredient"}
	createdIngredient, err := repo.Create(ctx, ingredient)

	// TODO make a direct Cypher query to verify anything about the state of the graph?

	assert := assert.New(t)
	assert.NoError(err)
	assert.NotEmpty(createdIngredient.Id)
	assert.Equal(createdIngredient.Name, ingredient.Name)
	assert.NotNil(createdIngredient.Created) // TODO more rigorous assertion of Created (e.g. same year, month, day. Or maybe with x time of now)?
	assert.Nil(createdIngredient.LastModified)
	assert.Nil(createdIngredient.Deleted)
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
