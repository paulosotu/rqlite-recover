package models

import (
	"flag"

	syncModel "github.com/paulosotu/local-storage-sync/models"
)

const (
	defaultHttpPort = "4001"
	defaultRaftPort = "4002"

	defaultLiveFilename  = "livez"
	defaultReadyFilename = "readyz"

	defaultNotReadyTimeoutSec = 180
	defaultNotAliveTimeoutSec = 60
)

type RecoverConfig struct {
	baseConfig    *syncModel.Config
	httpPort      string
	raftPort      string
	liveFilename  string
	readyFilename string

	notReadyTimeoutSec uint64
	notAliveTimeoutSec uint64
}

func NewRecoverConfigFromArgs() *RecoverConfig {

	ret := &RecoverConfig{}

	flag.StringVar(&ret.httpPort, "http", defaultHttpPort, "The http port for the rqlite api")
	flag.StringVar(&ret.raftPort, "raft", defaultRaftPort, "The raft port of rqlite")
	flag.StringVar(&ret.liveFilename, "live", defaultLiveFilename, "The name of the live filename")
	flag.StringVar(&ret.readyFilename, "ready", defaultReadyFilename, "The name of the ready filename")
	flag.Uint64Var(&ret.notAliveTimeoutSec, "ltime", defaultNotAliveTimeoutSec, "The number of seconds to wait before marking rqlite node as not alive.")
	flag.Uint64Var(&ret.notReadyTimeoutSec, "rtime", defaultNotReadyTimeoutSec, "The number of seconds to wait before marking rqlite node as not ready.")

	ret.baseConfig = syncModel.NewConfigFromArgs()

	return ret
}

func (c *RecoverConfig) GetHttpPort() string {
	return c.httpPort
}

func (c *RecoverConfig) GetRaftPort() string {
	return c.raftPort
}

func (c *RecoverConfig) GetLiveFilename() string {
	return c.liveFilename
}

func (c *RecoverConfig) GetReadyFilename() string {
	return c.readyFilename
}

func (c *RecoverConfig) GetReadinessTimeoutSec() uint64 {
	return c.notReadyTimeoutSec
}

func (c *RecoverConfig) GetLivenessTimeoutSec() uint64 {
	return c.notAliveTimeoutSec
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
