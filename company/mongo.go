package company

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collection = "companies"
)

type Mongo struct {
	db *mongo.Database
}

func NewMongo(db *mongo.Database) *Mongo {
	return &Mongo{
		db: db,
	}
}

func (m Mongo) Create(ctx context.Context, request Company) (Company, error) {
	_, err := m.db.Collection(collection).InsertOne(ctx, request)
	switch {
	case mongo.IsDuplicateKeyError(err):
		return Company{}, ErrDuplicatedEntry
	case err != nil:
		return Company{}, err
	}

	return request, nil
}

func (m Mongo) Fetch(ctx context.Context, lookup Lookup) ([]Company, error) {
	filter := make(bson.M)

	if lookup.ID != uuid.Nil {
		filter["_id"] = lookup.ID
	}
	cur, err := m.db.Collection(collection).Find(
		ctx,
		filter,
		&options.FindOptions{
			Sort: bson.M{"_id": 1},
		},
	)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var companies []Company
	for cur.Next(ctx) {
		var c Company
		err := cur.Decode(&c)
		if err != nil {
			return nil, err
		}
		companies = append(companies, c)
	}

	return companies, nil
}

func (m Mongo) FetchOne(ctx context.Context, lookup Lookup) (Company, error) {
	companies, err := m.Fetch(ctx, lookup)
	if err != nil {
		return Company{}, err
	}
	if len(companies) == 0 {
		return Company{}, errors.New("not found")
	}
	if len(companies) > 1 {
		return Company{}, errors.New("unexpected result")
	}
	return companies[0], nil
}

func (m Mongo) UpdateOne(ctx context.Context, lookup Lookup, request Company) (Company, error) {
	update := bson.M{
		"$set": request,
	}
	filter := make(bson.M)
	if lookup.ID != uuid.Nil {
		filter["_id"] = lookup.ID
	}
	res := m.db.Collection(collection).FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		return Company{}, res.Err()
	}

	var c Company
	err := res.Decode(&c)
	if err != nil {
		return Company{}, err
	}
	return c, nil
}

func (m Mongo) DeleteOne(ctx context.Context, lookup Lookup) error {
	filter := make(bson.M)
	if lookup.ID != uuid.Nil {
		filter["_id"] = lookup.ID
	}
	_, err := m.db.Collection(collection).DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
