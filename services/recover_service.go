package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/paulosotu/local-storage-sync/models"
	"github.com/paulosotu/local-storage-sync/services"
	log "github.com/sirupsen/logrus"
)

type RecoverService struct {
	ticker          *time.Ticker
	storageLocation services.StorageLocationService

	currentNodeName string
	rootDataPath    string
}

type Recover struct {
	Id       string `json:"id"`
	Address  string `json:"address"`
	NonVoter bool   `json:"non_voter"`
}

func NewRecoverService(tick int, storage services.StorageLocationService, nodeName, dataPath string) *RecoverService {
	return &RecoverService{
		ticker:          time.NewTicker(time.Duration(tick) * time.Second),
		storageLocation: storage,
		currentNodeName: nodeName,
		rootDataPath:    dataPath,
	}
}

func (r *RecoverService) Start(ctx context.Context) {
	log.Info("starting ...")
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-r.ticker.C:
				log.Info("coiso")
				storageLocations, err := r.storageLocation.GetStorageLocations()
				if err != nil {
					log.Errorf("Failed to get Storage Locations with error %v", err)
					continue
				}
				nodes, err := r.storageLocation.GetNodes()
				if err != nil {
					log.Errorf("Failed to get deamon set nodes with error %v", err)
					continue
				}
				locationsToRecover := make([]models.StoragePodLocation, 0, len(storageLocations))
				for _, stLocation := range storageLocations {
					if stLocation.GetNodeName() == r.currentNodeName {
						locationsToRecover = append(locationsToRecover, stLocation)
					}
				}

				log.Debugf("nodes: %v", nodes)
				log.Debugf("locations to recover: %v", locationsToRecover)
				log.Debugf("storageLocations: %v", storageLocations)
				if len(locationsToRecover) > 0 {
					r.handleRecovery(locationsToRecover, nodes)
				}
				continue
			}
		}
	}()
}

func (r *RecoverService) Stop() {
	r.ticker.Stop()
}

func buildRecoverData(nodes []models.Node) []Recover {
	recoverData := make([]Recover, 0, len(nodes))
	for _, node := range nodes {
		data := Recover{
			Id:       node.GetName(),
			Address:  node.GetIP(),
			NonVoter: false,
		}
		recoverData = append(recoverData, data)
	}
	return recoverData
}

func shouldWriteRecoverData(path string, newRecoverData []Recover) (bool, error) {
	var currentRecoverData []Recover

	filename := filepath.Join(path, "peers.json")

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		filename = filepath.Join(path, "peers.info")
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Debug("No peers file found, marking to should write!!")
			return true, nil
		}
	}
	recoverJson, err := os.Open(filename)

	if err != nil {
		return true, err
	}
	defer recoverJson.Close()
	recoverRaw, err := ioutil.ReadAll(recoverJson)

	if err != nil {
		return true, err
	}

	if err := json.Unmarshal(recoverRaw, &currentRecoverData); err != nil {
		return true, err
	}

	currentRecMap := make(map[string]Recover)

	for _, rec := range currentRecoverData {
		currentRecMap[rec.Id] = rec
	}

	for _, rec := range newRecoverData {

		if oldRec, ok := currentRecMap[rec.Id]; !ok {
			return true, nil
		} else if rec.Address != oldRec.Address || rec.NonVoter != oldRec.NonVoter {
			return true, nil
		}
	}

	return false, nil
}

func (r *RecoverService) handleRecovery(syncList []models.StoragePodLocation, nodes []models.Node) {
	root := r.rootDataPath

	if !filepath.IsAbs(r.rootDataPath) {
		root = filepath.Join("/", r.rootDataPath)
	}

	//Call this for each node that is not self
	for _, st := range syncList {
		path := root
		path = filepath.Join(path, st.GetHostDataDirName(), "file", "data", "raft")
		recoverData := buildRecoverData(nodes)

		log.Debugf("PATH: %v", path)
		log.Debugf("Recover: %v", nodes)

		if shouldWrite, err := shouldWriteRecoverData(path, recoverData); shouldWrite {
			if err != nil {
				log.Errorf("Will re-write recover file due to error %v", err)
			}
			b, err := json.Marshal(recoverData)
			if err != nil {
				log.Errorf("Failed to convert recover data to json with error %v", err)
				continue
			}
			if err = os.WriteFile(filepath.Join(path, "peers.json"), b, 0644); err != nil {
				log.Errorf("Failed to write peers.json due to error %v", err)
			}
			log.Debugf(string(b))
		}

	}
}
