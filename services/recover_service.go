package services

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
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

	podFailReadiness map[string]time.Time

	httpPort string
	raftPort string

	notReadyTimoutMin uint64
}

type Recover struct {
	Id       string `json:"id"`
	Address  string `json:"address"`
	NonVoter bool   `json:"non_voter"`
}

type RecoverConfig interface {
	GetTimerTick() int
	GetNodeName() string
	GetDataDir() string
	GetHttpPort() string
	GetRaftPort() string
	GetReadinessTimeoutMin() uint64
}

func NewRecoverService(cfg RecoverConfig, storage services.StorageLocationService) *RecoverService {
	return &RecoverService{
		ticker:            time.NewTicker(time.Duration(cfg.GetTimerTick()) * time.Second),
		storageLocation:   storage,
		currentNodeName:   cfg.GetNodeName(),
		rootDataPath:      cfg.GetDataDir(),
		podFailReadiness:  make(map[string]time.Time),
		httpPort:          cfg.GetHttpPort(),
		raftPort:          cfg.GetRaftPort(),
		notReadyTimoutMin: cfg.GetReadinessTimeoutMin(),
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
				storageLocations, err := r.storageLocation.GetStorageLocations()
				if err != nil {
					log.Errorf("Failed to get Storage Locations with error %v", err)
					continue
				}
				locationsToRecover := make([]models.StoragePodLocation, 0, len(storageLocations))
				for _, stLocation := range storageLocations {
					if stLocation.GetNodeName() == r.currentNodeName {
						locationsToRecover = append(locationsToRecover, stLocation)
					}
				}
				log.Debugf("locations to recover: %v", locationsToRecover)
				log.Debugf("storageLocations: %v", storageLocations)
				if len(locationsToRecover) > 0 {
					r.handleRecovery(ctx, locationsToRecover)
				}
				continue
			}
		}
	}()
}

func (r *RecoverService) Stop() {
	r.ticker.Stop()
}

func (r *RecoverService) buildRecoverData(stLocs []models.StoragePodLocation) []Recover {
	recoverData := make([]Recover, 0, len(stLocs))
	for _, loc := range stLocs {
		data := Recover{
			Id:       loc.GetBindPodName(),
			Address:  loc.GetPodIp() + ":" + r.raftPort,
			NonVoter: false,
		}
		recoverData = append(recoverData, data)

		if _, ok := r.podFailReadiness[data.Id]; !ok {
			r.podFailReadiness[data.Id] = time.Unix(0, 0)
		}
	}
	return recoverData
}

func (r *RecoverService) shouldWriteRecoverData(path string, newRecoverData []Recover) (bool, error) {
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

func (r *RecoverService) handleReadiness(ctx context.Context, podId, podIp, httpPort, podDataPath string) error {
	var req *http.Request
	var resp *http.Response
	var err error

	if req, err = http.NewRequestWithContext(ctx, "GET", "http://"+podIp+":"+httpPort+"/readyz", nil); err != nil {
		log.Errorf("Failed to check readiness due to error on building request: %v", err)
	} else {
		resp, err = http.DefaultClient.Do(req)
	}

	log.Debugf("Checking readiness with request %v", req)

	if err != nil || resp.StatusCode != 200 {
		if err != nil {
			log.Warnf("Not ready due to error: %v, %v", err)
		} else {
			log.Debugf("RQlite node responded NOT ready: %v, %v", err, resp.StatusCode)
		}
		if r.podFailReadiness[podId] == time.Unix(0, 0) {
			r.podFailReadiness[podId] = time.Now()
			return nil
		} else if uint64(math.Ceil(time.Since(r.podFailReadiness[podId]).Minutes())) > r.notReadyTimoutMin {
			err = os.Remove(filepath.Join(podDataPath, "readyz"))
			if err != nil {
				log.Errorf("Failed to remove readyz due to error %v", err)
				return err
			}
		}
	} else {
		log.Debugf("RQlite node responded ready: %v", resp.Body)
		r.podFailReadiness[podId] = time.Unix(0, 0)
	}
	return nil
}

func (r *RecoverService) handleRecovery(ctx context.Context, syncList []models.StoragePodLocation) {
	root := r.rootDataPath

	if !filepath.IsAbs(r.rootDataPath) {
		root = filepath.Join("/", r.rootDataPath)
	}

	//Call this for each node that is not self
	for _, st := range syncList {
		path := root
		path = filepath.Join(path, st.GetHostDataDirName(), "file", "data", "raft")
		recoverData := r.buildRecoverData(syncList)

		log.Debugf("PATH: %v", path)

		if shouldWrite, err := r.shouldWriteRecoverData(path, recoverData); shouldWrite {
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
			} else {
				log.Debugf(string(b))
				log.Infof("Wrote peers.json to %v", path)
			}

		}
		r.handleReadiness(ctx, st.GetBindPodName(), st.GetPodIp(), r.httpPort, path)
	}
}
