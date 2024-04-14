package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"
	"url-shortener/internal/cache"
	"url-shortener/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	records Records
	cache   cache.Cache
}

type Records struct {
	*mongo.Collection
}

type Record struct {
	Username string `bson:"username"`
	Alias    string `bson:"alias"`
	Url      string `bson:"url"`
}

func MustNew(timeout time.Duration, c cache.Cache, connString string, dbName string, collectionName string) *Store {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	newFunc := func() *Store {
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(connString))
		if err != nil {
			panic(err)
		}

		if err := client.Ping(ctx, nil); err != nil {
			panic(err)
		}

		records := Records{
			Collection: client.Database(dbName).Collection(collectionName),
		}

		return &Store{
			records: records,
			cache:   c,
		}
	}

	return newFunc()
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

func (s *Store) SaveURL(ctx context.Context, url, alias, username string) error {
	const op = "mongodb.SaveURL"

	filter := bson.D{{Key: "username", Value: username}, {Key: "alias", Value: alias}}

	var result Record
	err := s.records.FindOne(ctx, filter).Decode(&result)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return storage.ErrAliasAlreadyExist
	}

	_, err = s.records.InsertOne(ctx, Record{
		Username: username,
		Alias:    alias,
		Url:      url,
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := s.cache.Set(ctx, url, alias, username); err != nil {
		return fmt.Errorf("%s: %w: %w", op, storage.ErrCacheSet, err)
	}

	return nil
}

func (s *Store) GetURL(ctx context.Context, username, alias string) (string, error) {
	const op = "mongodb.GetURL"

	var returningErr error = fmt.Errorf("%s: ", op)

	if url, err := s.cache.Get(ctx, username, alias); err == nil {
		return url, nil
	} else {
		returningErr = fmt.Errorf("%w: %w: %w", returningErr, storage.ErrCacheGet, err)
	}

	filter := bson.D{{Key: "username", Value: username}, {Key: "alias", Value: alias}}

	var result Record
	err := s.records.FindOne(ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return "", storage.ErrAliasNotFound
	} else if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if errors.Is(returningErr, storage.ErrCacheGet) {
		return result.Url, returningErr
	}

	return result.Url, nil
}

func (s *Store) DeleteURL(ctx context.Context, username, alias string) error {
	const op = "mongodb.DeleteURL"

	filter := bson.D{{Key: "username", Value: username}, {Key: "alias", Value: alias}}

	res, err := s.records.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if res.DeletedCount == 0 {
		return storage.ErrAliasNotFound
	}

	if err := s.cache.Delete(ctx, username, alias); err != nil {
		return fmt.Errorf("%s: %w: %w", op, storage.ErrCacheDelete, err)
	}

	return nil
}

func (s *Store) UpdateAlias(ctx context.Context, username, oldAlias, newAlias string) error {
	const op = "mongodb.UpdateAlias"

	filter := bson.D{{Key: "username", Value: username}, {Key: "alias", Value: newAlias}}

	var result Record
	err := s.records.FindOne(ctx, filter).Decode(&result)
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return storage.ErrNewAliasAlreadyExists
	}

	filter = bson.D{{Key: "username", Value: username}, {Key: "alias", Value: oldAlias}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "username", Value: username}, {Key: "alias", Value: newAlias}}}}

	res, err := s.records.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("%s: failed to update alias", op)
	}

	if res.ModifiedCount == 0 {
		return storage.ErrAliasNotFound
	}

	if err := s.cache.Update(ctx, username, oldAlias, newAlias); err != nil {
		return fmt.Errorf("%s: %w: %w", op, storage.ErrCacheUpdate, err)
	}

	return nil
}
