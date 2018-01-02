package main

import (
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	fsgConfig "github.com/hyperledger/fabric-sdk-go/pkg/config"
	bccspFactory "github.com/hyperledger/fabric/bccsp/factory"
	fcutil "github.com/hyperledger/fabric-sdk-go/pkg/fabric-txn/internal/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"fmt"
	"os"
)

// FabricSetup implementation
type BaseSetupImpl struct {
	Client          fab.FabricClient
	Channel         fab.Channel
	EventHub        fab.EventHub
	Initialized      bool
	ChannelId        string
	ChannelConfig    string
	ChaincodeId      string
	ChaincodeVersion string
	ChaincodeGoPath  string
	ChaincodePath    string
	AdminUser       ca.User
}

// Initialize reads the configuration file and sets up the client, chain and event hub
func (setup *BaseSetupImpl) Initialize() () {

	// Create SDK setup for the integration tests
	sdkOptions := deffab.Options{
		ConfigFile: setup.ConfigFile,
	}

	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return errors.WithMessage(err, "SDK init failed")
	}
}

// getEventHub initialize the event hub
func getEventHub(client api.FabricClient) (api.EventHub, error) {
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, fmt.Errorf("Error creating new event hub: %v", err)
	}
	foundEventHub := false
	peerConfig, err := client.GetConfig().GetPeersConfig()
	if err != nil {
		return nil, fmt.Errorf("Error reading peer config: %v", err)
	}
	for _, p := range peerConfig {
		if p.EventHost != "" && p.EventPort != 0 {
			fmt.Printf("EventHub connect to peer (%s:%d)\n", p.EventHost, p.EventPort)
			eventHub.SetPeerAddr(fmt.Sprintf("%s:%d", p.EventHost, p.EventPort),
				p.TLS.Certificate, p.TLS.ServerHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, fmt.Errorf("No EventHub configuration found")
	}

	return eventHub, nil
}

// Install and instantiate the chaincode
func (setup *FabricSetup) InstallAndInstantiateCC() error {

	// Check if chaincode ID is provided
	// otherwise, generate a random one
	if setup.ChaincodeId == "" {
		setup.ChaincodeId = fcutil.GenerateRandomID()
	}

	fmt.Printf(
		"Chaincode %s (version %s) will be installed (Go Path: %s / Chaincode Path: %s)\n",
		setup.ChaincodeId,
		setup.ChaincodeVersion,
		setup.ChaincodeGoPath,
		setup.ChaincodePath,
	)

	// Install ChainCode
	// Package the go code and make a proposal to the network with this new chaincode
	err := fcutil.SendInstallCC(
		setup.Client, // The SDK client
		setup.Channel, // The channel concerned
		setup.ChaincodeId,
		setup.ChaincodePath,
		setup.ChaincodeVersion,
		nil,
		setup.Channel.GetPeers(), // Peers concerned by this change in the channel
		setup.ChaincodeGoPath,
	)
	if err != nil {
		return fmt.Errorf("Send install proposal return error: %v", err)
	} else {
		fmt.Printf("Chaincode %s installed (version %s)\n", setup.ChaincodeId, setup.ChaincodeVersion)
	}

	// Instantiate ChainCode
	// Call the Init function of the chaincode in order to initialize in every peer the new chaincode
	err = fcutil.SendInstantiateCC(
		setup.Channel,
		setup.ChaincodeId,
		setup.ChannelId,
		[]string{"init"}, // Arguments for the invoke request
		setup.ChaincodePath,
		setup.ChaincodeVersion,
		[]api.Peer{setup.Channel.GetPrimaryPeer()}, // Which peer to contact
		setup.EventHub,
	)
	if err != nil {
		return err
	} else {
		fmt.Printf("Chaincode %s instantiated (version %s)\n", setup.ChaincodeId, setup.ChaincodeVersion)
	}

	return nil
}
