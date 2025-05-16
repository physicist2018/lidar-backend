package licelfile

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/physicist2018/lidar-backend/internal/db/mongocon"
	"github.com/physicist2018/lidar-backend/internal/domain/lidardb"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type licelRepository struct {
	db *db.Mongo
}

func NewLicelRepository(db *db.Mongo) *licelRepository {
	return &licelRepository{db: db}
}

// CreateExperiment- в БД создает новый эксперимент (если его нет) или возвращает существующий
func (r *licelRepository) CreateExperiment(startTime time.Time, title, comments string) (*lidardb.Experiment, error) {
	exp := lidardb.NewExperiment(startTime, title, comments)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Производим вставку эксперимента в коллекцию Experiments
	_, err := r.db.Experiments.InsertOne(ctx, exp)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Info().Err(err).Send()
			return nil, fmt.Errorf("эксперимент с таким временем начала уже существует")
		}
		log.Error().Err(err).Send()
		return nil, fmt.Errorf("ошибка при сохранении эксперимента: %w", err)
	}

	// Возвращаем сохраненный эксперимент
	return exp, nil
}

func (r *licelRepository) FindExperimentByTime(startTime time.Time) (*lidardb.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exp lidardb.Experiment
	err := r.db.Experiments.FindOne(ctx, bson.M{"start_time": startTime}).Decode(&exp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Info().Err(err).Send()
			return nil, nil // Эксперимент не найден, но это не ошибка
		}
		log.Error().Err(err).Send()
		return nil, err
	}
	return &exp, nil
}

func (r *licelRepository) FindExperimentByTimeAndTitle(startTime time.Time, title string) (*lidardb.Experiment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exp lidardb.Experiment
	err := r.db.Experiments.FindOne(ctx, bson.M{"start_time": startTime, "title": title}).Decode(&exp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Info().Err(err).Send()
			return nil, nil // Эксперимент не найден, но это не ошибка
		}
		log.Error().Err(err).Send()
		return nil, err
	}
	return &exp, nil
}

func (r *licelRepository) FindMeasurement(expID bson.ObjectID, startTime, stopTime time.Time) (*lidardb.Measurement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var meas lidardb.Measurement
	err := r.db.Measurements.FindOne(ctx, bson.M{"experiment_id": expID, "data.measurementstarttime": startTime, "data.measurementstoptime": stopTime}).Decode(&meas)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Info().Err(err).Send()
			return nil, nil // Специальный случай: измерение не найдено, но это не ошибка
		}
		log.Error().Err(err).Send()
		return nil, err
	}
	return &meas, nil
}

func (r *licelRepository) CreateMeasurement(expId bson.ObjectID, meas *lidardb.Measurement) (*lidardb.Measurement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Устанавливаем связь измерения с экспериментом
	meas.ExperimentID = expId

	// Выполняем вставку в базу данных
	_, err := r.db.Measurements.InsertOne(ctx, meas)
	if err != nil {
		// Передаём ошибку вызывающему коду, чтобы тот мог корректно обработать ситуацию
		return nil, fmt.Errorf("ошибка при вставке измерения: %w", err)
	}

	// Возвращаем созданное измерение и nil в качестве ошибки
	return meas, nil
}

func (r *licelRepository) FindProcessingTask(expID bson.ObjectID) (*lidardb.ProcessingTask, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var task lidardb.ProcessingTask
	err := r.db.Processing.FindOne(ctx, bson.M{"experiment_id": expID}).Decode(&task)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Info().Err(err).Send()
			return nil, nil // Обработка случая, когда задача не найдена
		}
		log.Error().Err(err).Send()
		return nil, err
	}
	return &task, nil
}

func (r *licelRepository) CreateProcessingTask(expID bson.ObjectID, procTask *lidardb.ProcessingTask) (*lidardb.ProcessingTask, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	procTask.ExperimentID = expID
	_, err := r.db.Processing.InsertOne(ctx, procTask)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании задачи обработки: %w", err)
	}
	return procTask, nil
}

func (r *licelRepository) FindAllMeasurementsByExpID(expID bson.ObjectID) ([]*lidardb.Measurement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var measurements []*lidardb.Measurement
	cursor, err := r.db.Measurements.Find(ctx, bson.M{"experiment_id": expID})
	if err != nil {
		log.Error().Err(err).Send()

		return nil, err
	}
	err = cursor.All(ctx, &measurements)
	if err != nil {
		log.Error().Err(err).Send()

		return nil, err
	}
	return measurements, nil
}

// func (r *licelRepository) InsertExperiment(expPack *licelformat.LicelPack, comments string) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()

// 	var docs []lidardb.Measurement

// 	startTime := time.Now()
// 	for _, value := range expPack.Data {
// 		if value.MeasurementStartTime.Before(startTime) {
// 			startTime = value.MeasurementStartTime
// 		}
// 	}

// 	exp := &lidardb.Experiment{
// 		StartTime: startTime,
// 		Comments:  comments,
// 	}

// 	res, err := r.db.Database.Collection("rawExperiments").InsertOne(ctx, exp)
// 	insertedID, ok := res.InsertedID.(bson.ObjectID)
// 	if !ok {
// 		return err
// 	}
// 	if err != nil {
// 		return err
// 	}

// 	for _, value := range expPack.Data {
// 		docs = append(docs, lidardb.Measurement{Data: value, ExperimentID: insertedID})
// 	}

// 	_, err = r.db.Database.Collection("rawMeasurements").InsertMany(ctx, docs)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *licelRepository) GetExperiment(hexID string) ([]lidardb.Measurement, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
// 	measurements := r.db.Database.Collection("rawMeasurements")
// 	experiments := r.db.Database.Collection("rawExperiments")

// 	expID, err := bson.ObjectIDFromHex(hexID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var doc1 lidardb.Experiment

// 	err = experiments.FindOne(ctx, bson.D{{"_id", expID}}).Decode(&doc1)

// 	if err != nil {
// 		log.Printf("%v", err)
// 		return nil, err
// 	}

// 	fmt.Println(doc1)

// 	var doc2 []lidardb.Measurement

// 	filter := bson.M{"experiment_id": doc1.ID}
// 	cursor, err := measurements.Find(ctx, filter)

// 	if err != nil {
// 		log.Fatalf("%v", err)
// 	}

// 	if err = cursor.All(ctx, &doc2); err != nil {
// 		return nil, err
// 	}

// 	return doc2, nil
// }

// func (r *licelRepository) AddNewFieldToDocument(hexIDStr string, name string, data interface{}) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	update := bson.D{
// 		{"$set", bson.D{
// 			{Key: name, Value: data},
// 		}},
// 	}
// 	id, _ := bson.ObjectIDFromHex(hexIDStr)
// 	_, err := r.db.Database.Collection("rawExperiments").UpdateOne(ctx, bson.D{{"_id", id}},
// 		update)
// 	return err
// }
