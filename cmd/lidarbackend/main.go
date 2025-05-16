package main

import (
	"github.com/physicist2018/licelfile/licelformat"
	"github.com/rs/zerolog/log"

	"github.com/physicist2018/lidar-backend/internal/db/mongocon"
	"github.com/physicist2018/lidar-backend/internal/domain/lidardb"
	"github.com/physicist2018/lidar-backend/internal/repository/licelfile"
	licelfile2 "github.com/physicist2018/lidar-backend/internal/services/licelfile"
)

func main() {

	mongoConn, err := mongocon.NewMongo("mongodb://root:example@localhost:27017", "lidardb")
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer mongoConn.Close()
	//mongoConn.PrepareRawExperimentCollection()

	repo := licelfile.NewLicelRepository(mongoConn)
	srv := licelfile2.NewLicelService(repo)
	data := licelformat.NewLicelPack("b*")

	exp, err := srv.AddExperiment(data.StartTime, "test", "---")
	if err != nil {
		log.Error().Err(err).Msg("Какая-то ошибка")
		return
	}

	for _, v := range data.Data {
		//fmt.Println(k, v)
		tmpMeas := lidardb.NewMeasurement(exp.ID, &v)
		_, err = srv.AddMeasurement(exp.ID, tmpMeas)
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
	}

	srv.CreateProcessingTask(exp.ID)

}
