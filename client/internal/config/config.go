package config

import (
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

type Config struct {
	GamePatchPath  string
	DataDir        string
	UpstreamServer string
	UpstreamPort   uint16
	Verbose        bool
	LogFile        string
}

func ParseConfig(args []string) (*Config, error) {
	var defaultDataDir string
	confDir, err := os.UserConfigDir()
	if err != nil {
		defaultDataDir = "./data/client"
	} else {
		defaultDataDir = confDir + "/OoTMM/client"
	}

	var conf Config
	fs := flag.NewFlagSet("client", flag.ContinueOnError)
	fs.StringVarP(&conf.DataDir, "data", "d", defaultDataDir, "Path to the data directory")
	fs.StringVarP(&conf.UpstreamServer, "server", "s", "multi.ootmm.com", "Upstream server address")
	fs.Uint16VarP(&conf.UpstreamPort, "port", "p", 14236, "Upstream server port")
	fs.BoolVarP(&conf.Verbose, "verbose", "v", false, "Print the messages sent to the daemon (the payloads broadcast over the websocket) to the console")
	fs.StringVar(&conf.LogFile, "log-file", "", "Append the messages sent to the daemon to this file (may contain session IDs and player names)")
	err = fs.Parse(args)
	if err != nil {
		return nil, err
	}

	posArgs := fs.Args()
	if len(posArgs) == 1 {
		conf.GamePatchPath = posArgs[0]
	} else {
		return nil, fmt.Errorf("expected exactly one positional argument (path to game patch), got %d", len(posArgs))
	}

	return &conf, nil
}
