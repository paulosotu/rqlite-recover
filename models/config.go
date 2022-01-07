package models

import (
	"flag"

	syncModel "github.com/paulosotu/local-storage-sync/models"
)

const (
	defaultHttpPort = "4001"
	defaultRaftPort = "4002"

	defaultNotReadyTimeoutMin = 3
)

type RecoverConfig struct {
	baseConfig *syncModel.Config
	httpPort   string
	raftPort   string

	NotReadyTimeoutMin uint64
}

func NewRecoverConfigFromArgs() *RecoverConfig {

	ret := &RecoverConfig{}

	flag.StringVar(&ret.httpPort, "http", defaultHttpPort, "The http port for the rqlite api")
	flag.StringVar(&ret.raftPort, "raft", defaultRaftPort, "The raft port of rqlite")
	flag.Uint64Var(&ret.NotReadyTimeoutMin, "timeout", defaultNotReadyTimeoutMin, "The number of minutes to wait before marking rqlite node as unready")

	ret.baseConfig = syncModel.NewConfigFromArgs()

	return ret
}

func (c *RecoverConfig) GetHttpPort() string {
	return c.httpPort
}

func (c *RecoverConfig) GetRaftPort() string {
	return c.raftPort
}

func (c *RecoverConfig) GetReadinessTimeoutMin() uint64 {
	return c.NotReadyTimeoutMin
}

func (c *RecoverConfig) GetTimerTick() int {
	return c.baseConfig.GetTimerTick()
}

func (c *RecoverConfig) GetNodeName() string {
	return c.baseConfig.GetNodeName()
}

func (c *RecoverConfig) GetDataDir() string {
	return c.baseConfig.GetDataDir()
}

func (c *RecoverConfig) GetDSAppName() string {
	return c.baseConfig.GetDSAppName()
}

func (c *RecoverConfig) GetSelector() string {
	return c.baseConfig.GetSelector()
}

func (c *RecoverConfig) GetExitCodeInterrupt() int {
	return c.baseConfig.GetExitCodeInterrupt()
}

func (c *RecoverConfig) ShouldRunInCluster() bool {
	return c.baseConfig.ShouldRunInCluster()
}

func (c *RecoverConfig) GetLogLevel() string {
	return c.baseConfig.GetLogLevel()
}
