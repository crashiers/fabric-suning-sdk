package main

import (
	"fmt"
	"os"
	"runtime"
	"path/filepath"
	"github.com/fabric-suning-sdk/blockchain"
	"github.com/fabric-suning-sdk/web"
	"github.com/fabric-suning-sdk/web/controllers"
)


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

	// Initialize the Fabric SDK
	fabricSdk, err := blockchain.Initialize()
	if err != nil {
		fmt.Printf("Unable to initialize the Fabric SDK: %v\n", err)
	}

	// Install and instantiate the chaincode
	err = fabricSdk.InstallAndInstantiateCC()
	if err != nil {
		fmt.Printf("Unable to install and instantiate the chaincode: %v\n", err)
	}

	// Make the web application listening
	app := &controllers.Application{
		Fabric: fabricSdk,
	}
	web.Serve(app)
}
