package main

import (
	"fmt"

	"os"
	"path/filepath"
	"runtime"
)

var base BaseSetupImpl

type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func defaultGOPATH() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	if home := os.Getenv(env); home != "" {
		def := filepath.Join(home, "go")
		if filepath.Clean(def) == filepath.Clean(runtime.GOROOT()) {
			// Don't set the default GOPATH to GOROOT,
			// as that will trigger warnings from the go tool.
			return ""
		}
		return def
	}
	return ""
}

func init() {
	// Setup correctly the GOPATH in the environment
	if goPath := os.Getenv("GOPATH"); goPath == "" {
		os.Setenv("GOPATH", defaultGOPATH())
	}

	base = BaseSetupImpl{
		ConfigFile:      "./config.yaml",
		ChannelID:       "mychannel",
		OrgID:           "Org1",
		ChannelConfig:   "./test/fixtures/fabric/channel-artifacts/channel.tx",
		ChainCodeID:     "suningCC",
		ConnectEventHub: true,
	}

	fmt.Printf("Start to Initialize the Fabric SDK")
	if err := base.Initialize(); err != nil {
		fmt.Printf("Initialize: %v", err)
		os.Exit(-1)
	}

	fmt.Printf("Start to install and instantiate the suning chaincode\n")
	if err := base.InstallAndInstantiateSuningCC(); err != nil {
		fmt.Printf("Install and instantiate the suning chaincode failed:%v", err)
	}

}

func main() {

}
