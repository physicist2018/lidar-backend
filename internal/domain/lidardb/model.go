package lidardb

import (
	"time"

	"github.com/physicist2018/licelfile/licelformat"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	Experiment struct {
		ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
		StartTime time.Time     `bson:"start_time" json:"start_time"`
		Title     string        `bson:"title" json:"title"`
		Comments  string        `bson:"comments" json:"comments"`
	}

	Measurement struct {
		ID           bson.ObjectID          `bson:"_id,omitempty" json:"id"`
		ExperimentID bson.ObjectID          `bson:"experiment_id" json:"experiment_id"`
		Data         *licelformat.LicelFile `bson:"data" json:"data"`
	}

	ProcessingTask struct {
		ID             bson.ObjectID          `bson:"_id,omitempty" json:"id"`
		ExperimentID   bson.ObjectID          `bson:"experiment_id" json:"experiment_id"`
		MeasurementIDs []bson.ObjectID        `bson:"measurement_ids" json:"measurement_ids"`
		AveragedData   *licelformat.LicelFile `bson:"averaged_data" json:"averaged_data"` // Как только начинаем обработку,
		//создаем новую записл, содержащую уследненные данные отдного экспериметнта
	}
)

// NewExperiment is a function that creates a new Experiment.
// It takes a start time and comments as arguments and returns a pointer to an Experiment.
// It creates a new Experiment with a new ID, the given start time, and the given comments.
func NewExperiment(startTime time.Time, title, comments string) *Experiment {
	return &Experiment{
		ID:        bson.NewObjectID(),
		StartTime: startTime,
		Title:     title,
		Comments:  comments,
	}
}

// NewMeasurement is a function that creates a new Measurement.
// It takes an experiment ID and data as arguments and returns a pointer to a Measurement.
// It creates a new Measurement with a new ID, the given experiment ID, and the given data.
func NewMeasurement(experimentID bson.ObjectID, data *licelformat.LicelFile) *Measurement {
	return &Measurement{
		ID:           bson.NewObjectID(),
		ExperimentID: experimentID,
		Data:         data,
	}
}

func NewProcessingTask(experimentID bson.ObjectID, data *licelformat.LicelFile) *ProcessingTask {
	return &ProcessingTask{
		ID:           bson.NewObjectID(),
		ExperimentID: experimentID,
		AveragedData: data,
	}
}

// вынести в пакет licelformat
// AverageMeasurements is a function that takes a slice of Measurements and returns an averaged LicelFile.
func AverageMeasurements(measurements []*Measurement) (*licelformat.LicelFile, error) {
	result := licelformat.LicelFile{}
	var stopTime time.Time
	startTime := time.Now()
	for _, measurement := range measurements {
		// TODO: Реализовать вычисление среднего значения
		//result = licelformat.LicelFile{} // Заменить на реальное вычисление среднего значения
		if measurement.Data.MeasurementStartTime.Before(startTime) {
			startTime = measurement.Data.MeasurementStartTime
		}
		if measurement.Data.MeasurementStopTime.After(stopTime) {
			stopTime = measurement.Data.MeasurementStopTime
		}

		if len(result.Profiles) == 0 {
			result.Profiles = measurement.Data.Profiles
			result.MeasurementSite = measurement.Data.MeasurementSite
			result.Laser1NShots = measurement.Data.Laser1NShots
			result.Laser1Freq = measurement.Data.Laser1Freq
			result.Laser2NShots = measurement.Data.Laser2NShots
			result.Laser2Freq = measurement.Data.Laser2Freq
			result.Laser3NShots = measurement.Data.Laser3NShots
			result.Laser3Freq = measurement.Data.Laser3Freq
		} else {
			result.Laser1NShots += measurement.Data.Laser1NShots
			result.Laser2NShots += measurement.Data.Laser2NShots
			result.Laser3NShots += measurement.Data.Laser3NShots
			for i := range result.Profiles {
				result.Profiles[i].NShots += measurement.Data.Profiles[i].NShots
				for j := range measurement.Data.Profiles[i].Data {
					result.Profiles[i].Data[j] += measurement.Data.Profiles[i].Data[j]
				}
			}
		}

	}
	result.MeasurementStartTime = startTime
	result.MeasurementStopTime = stopTime
	return &result, nil
}
