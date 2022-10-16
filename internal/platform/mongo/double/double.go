package double

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"go.mongodb.org/mongo-driver/bson"
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
			Cmd:        []string{"mongod", "--replSet", "rs", "--keyFile", "/etc/mongo/mongo-repl.key"},
			Entrypoint: []string{
				"bash", "-c", "mkdir /etc/mongo\n" +
					"openssl rand -base64 768 > /etc/mongo/mongo-repl.key\n" +
					"chmod 400 /etc/mongo/mongo-repl.key\n" +
					"chown 999:999 /etc/mongo/mongo-repl.key\n" +
					"exec docker-entrypoint.sh $@"},
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
		res := client.Database("admin").RunCommand(ctx, bson.M{"replSetInitiate": bson.M{
			"_id": "rs",
			"members": []bson.M{
				{"_id": 0, "host": "localhost:27017"},
			},
		}})
		if err := res.Err(); err != nil {
			fmt.Println(err)
			return err
		}

		var repl struct {
			State int `bson:"myState"`
		}
		// Wait for RS to initialize.
		for repl.State != 1 {
			time.Sleep(100 * time.Millisecond)
			res = client.Database("admin").RunCommand(ctx, bson.M{"replSetGetStatus": 1})
			if err := res.Err(); err != nil {
				return err
			}
			_ = res.Decode(&repl)
		}

		return nil
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
