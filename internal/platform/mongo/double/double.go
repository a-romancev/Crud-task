package double

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewDocker() *mongo.Database {
	return clientFromContainer().Database(uuid.NewString())
}

func Purge() {
	m.Lock()
	defer m.Unlock()

	if purge != nil {
		purge()
	}
}

var (
	m      sync.Mutex
	client *mongo.Client
	purge  func()
)

func clientFromContainer() *mongo.Client {
	m.Lock()
	defer m.Unlock()

	if client != nil {
		return client
	}

	ctx := context.Background()

	pool, err := dockertest.NewPool(os.Getenv("DOCKER_HOST"))
	if err != nil {
		panic(err)
	}

	resource, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository: "mongo",
			Tag:        "4.4",
			Env: []string{
				"MONGO_INITDB_ROOT_USERNAME=mongo",
				"MONGO_INITDB_ROOT_PASSWORD=mongo",
			},
		},
		func(hc *docker.HostConfig) {
			hc.AutoRemove = true
			hc.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := recover(); err != nil {
			_ = pool.Purge(resource)
			panic(err)
		}
	}()

	if err := resource.Expire(60); err != nil {
		panic(err)
	}

	err = pool.Retry(func() error {
		port := resource.GetPort("27017/tcp")
		client, err = mongo.Connect(ctx, options.Client().ApplyURI(fmt.Sprintf("mongodb://mongo:mongo@localhost:%s", port)).SetDirect(true))
		if err != nil {
			return err
		}
		return client.Ping(ctx, nil)
	})
	if err != nil {
		panic(err)
	}

	purge = func() {
		_ = client.Disconnect(ctx)
		_ = pool.Purge(resource)
	}

	return client
}