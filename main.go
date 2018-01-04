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

func main() {
	// Setup correctly the GOPATH in the environment
	if goPath := os.Getenv("GOPATH"); goPath == "" {
		os.Setenv("GOPATH", defaultGOPATH())
	}

	base = BaseSetupImpl{
		ConfigFile:      "./config.yaml",
		ChannelID:       "testchannel1",
		OrgID:           "Org1",
		ChannelConfig:   "../test/fixtures/fabric/channel-artifacts/testchannel.tx",
		ChainCodeID:     "testcc",
		ConnectEventHub: true,
	}

	// Initialize the Fabric SDK
	if err := base.Initialize(); err != nil {
		fmt.Printf("Initialize: %v", err)
		os.Exit(-1)
	}

	// Install and instantiate the chaincode
	//	err = fabricSdk.InstallAndInstantiateCC()
	//	if err != nil {
	//		fmt.Printf("Unable to install and instantiate the chaincode: %v\n", err)
	//	}

	//	// Make the web application listening
	//	app := &controllers.Application{
	//		Fabric: fabricSdk,
	//	}
	//	web.Serve(app)
}
