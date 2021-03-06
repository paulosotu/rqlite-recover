package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	lssServices "github.com/paulosotu/local-storage-sync/pkg/services"

	"github.com/paulosotu/rqlite-recover/pkg/models"
	services "github.com/paulosotu/rqlite-recover/pkg/services"
	log "github.com/sirupsen/logrus"
)

type ContextKey string

const CONFIG_KEY = ContextKey("config")

func initLogger(logLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)

	}
}

func main() {
	config := models.NewRecoverConfigFromArgs()

	initLogger(config.GetLogLevel())

	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), CONFIG_KEY, *config))

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)

	go func() {
		select {
		case sig := <-sigs: // first signal, cancel context
			cancel()
			log.Infof("Received sigterm, exiting gracefully %s\n", sig.String())
		case <-ctx.Done():
			return
		}
		<-sigs // second signal, hard exit
		os.Exit(config.GetExitCodeInterrupt())

	}()

	defer func() { // all the cleanup here please
		cancel()
		close(sigs)
	}()

	if err := run(ctx); err != nil {
		log.Fatalf("[main] Fatal error: %s", err)
	}
}

func run(ctx context.Context) error {
	knService, err := startKubeCorePVCService(ctx)
	if err != nil {
		return err
	}
	recoverService, err := startRecoverService(ctx, knService)
	if err != nil {
		return err
	}

	defer func() { // all the cleanup here please
		knService.Stop()
		recoverService.Stop()
	}()

	<-ctx.Done()

	log.Info("leaving main loop...")
	return err
}

func startRecoverService(ctx context.Context, storageService lssServices.StorageLocationService) (*services.RecoverService, error) {
	config := ctx.Value(CONFIG_KEY).(models.RecoverConfig)

	serv := services.NewRecoverService(&config, storageService)
	serv.Start(ctx)
	return serv, nil
}

func startKubeCorePVCService(ctx context.Context) (*lssServices.KubeCorePVCService, error) {
	config := ctx.Value(CONFIG_KEY).(models.RecoverConfig)

	knService := lssServices.NewKubeCorePVCService(&config)
	knService.Start()

	log.Info("Waiting For KubeCorePVCService readiness!")
	if knService.WaitForReady(ctx) {
		log.Info("KubeCorePVCService is ready")

		list, err := knService.GetStorageLocations()
		if err != nil {
			log.Errorf("failed to get list: %s\n", err)
			return nil, err
		}
		fmt.Printf("%-32s%-45s%-28s%-22s%-22s%-8s\n", "PVC", "Volume", "POD", "Node", "namespace", "Status")
		for _, val := range list {
			fmt.Printf("%s\n", val.String())
		}
	}
	return knService, nil
}
