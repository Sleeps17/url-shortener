package mongodb

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"url-shortener/internal/storage"
)

type Store struct {
	records Records
}

type Records struct {
	*mongo.Collection
}

type Record struct {
	Alias string `bson:"alias"`
	Url   string `bson:"url"`
}

func MustNew(connectionString string, timeout time.Duration) *Store {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		panic(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		panic(err)
	}

	records := Records{
		Collection: client.Database("url-shortener").Collection("urls"),
	}

	return &Store{
		records: records,
	}
}

func (s *Store) Close(ctx context.Context) error {
	return s.records.Database().Client().Disconnect(ctx)
}

func (s *Store) SaveURL(ctx context.Context, url, alias string) error {
	const op = "mongodb.SaveURL"

	filter := bson.D{{"alias", alias}}

	var result Record
	err := s.records.FindOne(ctx, filter).Decode(&result)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return storage.ErrAliasAlreadyExist
	}

	_, err = s.records.InsertOne(ctx, Record{
		Alias: alias,
		Url:   url,
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Store) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "mongodb.GetURL"

	filter := bson.D{{"alias", alias}}

	var result Record
	err := s.records.FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", storage.ErrAliasNotFound
	} else if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return result.Url, nil
}

func (s *Store) DeleteURL(ctx context.Context, alias string) error {
	const op = "mongodb.DeleteURL"

	filter := bson.D{{"alias", alias}}

	res, err := s.records.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.DeletedCount == 0 {
		return storage.ErrAliasNotFound
	}

	return nil
}

func (s *Store) UpdateAlias(ctx context.Context, oldAlias, newAlias string) error {
	const op = "mongodb.UpdateAlias"

	filter := bson.D{{"alias", newAlias}}

	var result Record
	err := s.records.FindOne(ctx, filter).Decode(&result)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return storage.ErrNewAliasAlreadyExists
	}

	filter = bson.D{{"alias", oldAlias}}
	update := bson.D{{"$set", bson.D{{"alias", newAlias}}}}

	res, err := s.records.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%s: failed to update alias", op)
	}

	if res.ModifiedCount == 0 {
		return storage.ErrAliasNotFound
	}

	return nil
}
