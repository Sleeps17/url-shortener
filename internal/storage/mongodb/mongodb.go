package mongodb

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"url-shortener/internal/cache"
	mapCache "url-shortener/internal/cache/map-cache"
	"url-shortener/internal/storage"
)

type Store struct {
	records Records
	cache   cache.Cache
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

	c, err := mapCache.New(30)
	if err != nil {
		panic(err)
	}

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
		cache:   c,
	}
}

func (s *Store) Close(ctx context.Context) error {

	err1 := s.cache.Close(ctx)
	err2 := s.records.Database().Client().Disconnect(ctx)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("%w && %w", err1, err2)
	} else if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	return nil
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

	if err := s.cache.Set(ctx, url, alias); err != nil {
		return storage.ErrCacheSet
	}

	return nil
}

func (s *Store) GetURL(ctx context.Context, alias string) (string, error) {
	const op = "mongodb.GetURL"

	if url, err := s.cache.Get(ctx, alias); err == nil {
		return url, nil
	}

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

	if err := s.cache.Delete(ctx, alias); err != nil {
		return storage.ErrCacheDelete
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

	if err := s.cache.Update(ctx, oldAlias, newAlias); err != nil {
		return storage.ErrCacheUpdate
	}

	return nil
}
