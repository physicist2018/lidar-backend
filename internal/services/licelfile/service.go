package licelfile

import (
	"fmt"
	"time"

	"github.com/physicist2018/lidar-backend/internal/domain/lidardb"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type LicelRepo interface {
	CreateExperiment(startTime time.Time, title, comments string) (*lidardb.Experiment, error) // В БД создает новый эксперимент (если его нет) или возвращает существующий
	FindExperimentByTime(startTime time.Time) (*lidardb.Experiment, error)
	FindExperimentByTimeAndTitle(startTime time.Time, title string) (*lidardb.Experiment, error)
	CreateMeasurement(expId bson.ObjectID, meas *lidardb.Measurement) (*lidardb.Measurement, error) // Создает новое измерение, если оно еще не создано
	FindMeasurement(expID bson.ObjectID, startTime, stopTime time.Time) (*lidardb.Measurement, error)
	FindProcessingTask(expID bson.ObjectID) (*lidardb.ProcessingTask, error)
	FindAllMeasurementsByExpID(expID bson.ObjectID) ([]*lidardb.Measurement, error) // Получает все измерения эксперимента по его ID
	CreateProcessingTask(expID bson.ObjectID, procTask *lidardb.ProcessingTask) (*lidardb.ProcessingTask, error)
}

type licelService struct {
	repo LicelRepo
}

func NewLicelService(r LicelRepo) *licelService {
	return &licelService{
		repo: r,
	}
}

func (s *licelService) AddExperiment(startTime time.Time, title, comments string) (*lidardb.Experiment, error) {
	// Пробуем найти существующий эксперимент по времени старта
	foundExp, err := s.repo.FindExperimentByTimeAndTitle(startTime, title)
	if err != nil {
		return nil, fmt.Errorf("ошибка при поиске эксперимента: %w", err)
	}

	// Если эксперимент уже существует, возвращаем его
	if foundExp != nil {
		log.Debug().Str("ID", foundExp.ID.Hex()).Msg("Эксперимент успешно найден")
		return foundExp, nil
	}

	// Иначе создадим новый эксперимент
	newExp, err := s.repo.CreateExperiment(startTime, title, comments)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании эксперимента: %w", err)
	}

	return newExp, nil
}

func (s *licelService) AddMeasurement(expId bson.ObjectID, meas *lidardb.Measurement) (*lidardb.Measurement, error) {
	// Сначала пробуем найти существующее измерение
	foundMeas, err := s.repo.FindMeasurement(expId, meas.Data.MeasurementStartTime, meas.Data.MeasurementStopTime)
	if err != nil {
		// Вернуть ошибку, если возникли проблемы при поиске
		return nil, fmt.Errorf("ошибка при поиске измерения: %w", err)
	}

	// Если измерение уже существует, возвращаем его
	if foundMeas != nil {
		log.Debug().Str("ID", meas.ID.Hex()).Msg("Измерение успешно найдено")
		return foundMeas, nil
	}

	// Если измерение не найдено, создаём новое
	createdMeas, err := s.repo.CreateMeasurement(expId, meas)
	if err != nil {
		// Вернуть ошибку, если возникли проблемы при создании
		return nil, fmt.Errorf("ошибка при создании измерения: %w", err)
	}

	return createdMeas, nil
}

func (s *licelService) CreateProcessingTask(expId bson.ObjectID) (*lidardb.ProcessingTask, error) {
	// Сначала пробуем найти существующую задачу обработки
	foundTask, err := s.repo.FindProcessingTask(expId)
	if err != nil {
		// Вернуть ошибку, если возникли проблемы при поиске
		return nil, fmt.Errorf("ошибка при поиске задачи обработки: %w", err)
	}

	// Если задача обработки уже существует, возвращаем её
	if foundTask != nil {
		log.Debug().Str("ID", foundTask.ID.Hex()).Msg("Задача обработки успешно найдена")
		return foundTask, nil
	}

	// Если задача обработки не найдена, создаём новую
	// Найдем все измерения, связанные с экспериментом
	meas, err := s.repo.FindAllMeasurementsByExpID(expId)
	if err != nil {
		return nil, err
	}

	lf, err := lidardb.AverageMeasurements(meas)
	if err != nil {
		log.Error().Err(err).Send()
		return nil, err
	}

	newProcTask := lidardb.NewProcessingTask(expId, lf)
	newTask, err := s.repo.CreateProcessingTask(expId, newProcTask)
	if err != nil {
		// Вернуть ошибку, если возникли проблемы при создании
		return nil, fmt.Errorf("ошибка при создании задачи обработки: %w", err)
	}

	return newTask, nil
}
