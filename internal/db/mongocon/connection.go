package mongocon

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Mongo struct {
	Client       *mongo.Client
	Database     *mongo.Database
	Experiments  *mongo.Collection // здесь будут все эксперименты
	Measurements *mongo.Collection // здесь отдельно будут все измерения
	Processing   *mongo.Collection // здесь будет все результаты обработки
}

func NewMongo(uri string, dbName string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Info().Msg("[Ok ] Подключение к MongoDB установлено")

	db := client.Database(dbName)

	expCol, err := createCollectionIfNotExists(ctx, db, "rawExperiments")
	if err != nil {
		return nil, err
	}
	measCol, err := createCollectionIfNotExists(ctx, db, "rawMeasurements")
	if err != nil {
		return nil, err
	}

	procCol, err := createCollectionIfNotExists(ctx, db, "procExperiments")
	if err != nil {
		return nil, err
	}

	log.Info().Msg("[Ok ] Коллекции созданы или уже существуют")
	m := &Mongo{
		Client:       client,
		Database:     db,
		Experiments:  expCol,
		Measurements: measCol,
		Processing:   procCol,
	}
	return m, nil
}

func (m *Mongo) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

// createCollectionIfNotExists - Создает коллекцию, если она не существует
func createCollectionIfNotExists(ctx context.Context, db *mongo.Database, collectionName string) (*mongo.Collection, error) {
	// Проверяем, существует ли коллекция
	collections, err := db.ListCollectionNames(ctx, bson.M{"name": collectionName})
	if err != nil {
		log.Error().Err(err).Msg("[Err] Ошибка запроса списка коллекций")
		return nil, err
	}

	if len(collections) == 0 {
		log.Info().Str("Коллекция не найдена. Создаю...", collectionName).Send()

		// Создаём пустую коллекцию
		err := db.CreateCollection(ctx, collectionName)
		if err != nil {
			log.Error().Err(err).Msg("[Err] Ошибка создания коллекции")
			return nil, err
		}
		log.Info().Str("[Ok ] Коллекция создана.", collectionName).Send()
	} else {
		log.Info().Str("[Inf] Коллекция уже существует.", collectionName).Send()
	}

	return db.Collection(collectionName), nil
}
