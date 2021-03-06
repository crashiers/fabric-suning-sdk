/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	ca "github.com/hyperledger/fabric-sdk-go/api/apifabca"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	chmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/chmgmtclient"
	resmgmt "github.com/hyperledger/fabric-sdk-go/api/apitxn/resmgmtclient"
	packager "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/ccpackager/gopackager"
	pb "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/peer"

	deffab "github.com/hyperledger/fabric-sdk-go/def/fabapi"
	"github.com/hyperledger/fabric-sdk-go/pkg/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/errors"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/events"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/orderer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/peer"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/common/cauthdsl"
)

// BaseSetupImpl implementation of BaseTestSetup
type BaseSetupImpl struct {
	Client          fab.FabricClient
	Channel         fab.Channel
	EventHub        fab.EventHub
	ConnectEventHub bool
	ConfigFile      string
	OrgID           string
	ChannelID       string
	ChainCodeID     string
	Initialized     bool
	ChannelConfig   string
	AdminUser       ca.User
}

const (
	org1 = "Org1"
	org2 = "Org2"
	org3 = "Org3"
	org4 = "Org4"
)

var org1ResMgmt resmgmt.ResourceMgmtClient
var org2ResMgmt resmgmt.ResourceMgmtClient
var org3ResMgmt resmgmt.ResourceMgmtClient
var org4ResMgmt resmgmt.ResourceMgmtClient

// Initialize reads configuration from file and sets up client, channel and event hub
func (setup *BaseSetupImpl) Initialize() error {
	// Create SDK setup for the integration tests
	sdkOptions := deffab.Options{
		ConfigFile: setup.ConfigFile,
	}

	sdk, err := deffab.NewSDK(sdkOptions)
	if err != nil {
		return errors.WithMessage(err, "SDK init failed")
	}

	session, err := sdk.NewPreEnrolledUserSession(setup.OrgID, "Admin")
	if err != nil {
		return errors.WithMessage(err, "failed getting admin user session for org")
	}

	sc, err := sdk.NewSystemClient(session)
	if err != nil {
		return errors.WithMessage(err, "NewSystemClient failed")
	}

	setup.Client = sc
	setup.AdminUser = session.Identity()

	channel, err := setup.GetChannel(setup.Client, setup.ChannelID, []string{setup.OrgID})
	if err != nil {
		return errors.Wrapf(err, "create channel (%s) failed: %v", setup.ChannelID)
	}
	setup.Channel = channel

	// Channel management client is responsible for managing channels (create/update)
	chMgmtClient, err := sdk.NewChannelMgmtClientWithOpts("Admin", &deffab.ChannelMgmtClientOpts{OrgName: "ordererorg"})
	if err != nil {
		fmt.Errorf("Failed to create new channel management client: %s", err)
	}

	// Check if primary peer has joined channel
	alreadyJoined, err := HasPrimaryPeerJoinedChannel(sc, channel)
	if err != nil {
		return errors.WithMessage(err, "failed while checking if primary peer has already joined channel")
	}

	if !alreadyJoined {

		// Channel config signing user (has to belong to one of channel orgs)
		org1Admin, err := sdk.NewPreEnrolledUser("Org1", "Admin")
		if err != nil {
			return errors.WithMessage(err, "failed getting Org1 admin user")
		}

		// Create channel (or update if it already exists)
		req := chmgmt.SaveChannelRequest{ChannelID: setup.ChannelID, ChannelConfig: setup.ChannelConfig, SigningUser: org1Admin}

		if err = chMgmtClient.SaveChannel(req); err != nil {
			return errors.WithMessage(err, "SaveChannel failed")
		}

		time.Sleep(time.Second * 3)

		// org1 resource management client
		org1ResMgmt, err = sdk.NewResourceMgmtClient("Admin")
		if err != nil {
			return errors.WithMessage(err, "org1 failed to create new resource management client")
		}

		if err = channel.Initialize(nil); err != nil {
			return errors.WithMessage(err, "channel init failed")
		}

		if err = org1ResMgmt.JoinChannel(setup.ChannelID); err != nil {
			return errors.WithMessage(err, "org1 joinChannel failed")
		}

		// org2 resource management client
		org2ResMgmt, err = sdk.NewResourceMgmtClientWithOpts("Admin", &deffab.ResourceMgmtClientOpts{OrgName: org2})
		if err != nil {
			return errors.WithMessage(err, "org2 failed to create new resource management client")
		}

		if err = org2ResMgmt.JoinChannel(setup.ChannelID); err != nil {
			return errors.WithMessage(err, "org2 joinChannel failed")
		}

		// org3 resource management client
		org3ResMgmt, err = sdk.NewResourceMgmtClientWithOpts("Admin", &deffab.ResourceMgmtClientOpts{OrgName: org3})
		if err != nil {
			return errors.WithMessage(err, "org3 failed to create new resource management client")
		}

		if err = org3ResMgmt.JoinChannel(setup.ChannelID); err != nil {
			return errors.WithMessage(err, "org3 joinChannel failed")
		}

		// org4 resource management client
		org4ResMgmt, err = sdk.NewResourceMgmtClientWithOpts("Admin", &deffab.ResourceMgmtClientOpts{OrgName: org4})
		if err != nil {
			return errors.WithMessage(err, "org4 failed to create new resource management client")
		}

		if err = org4ResMgmt.JoinChannel(setup.ChannelID); err != nil {
			return errors.WithMessage(err, "org4 joinChannel failed")
		}

		fmt.Printf("Start to install and instantiate the suning chaincode\n")
		if err := base.InstallAndInstantiateSuningCC(); err != nil {
			fmt.Printf("Install and instantiate the suning chaincode failed:%v", err)
		}
	}

	if err := setup.setupEventHub(sc); err != nil {
		return err
	}

	setup.Initialized = true

	return nil
}

func (setup *BaseSetupImpl) setupEventHub(client fab.FabricClient) error {
	eventHub, err := setup.getEventHub(client)
	if err != nil {
		return err
	}

	if setup.ConnectEventHub {
		if err := eventHub.Connect(); err != nil {
			return errors.WithMessage(err, "eventHub connect failed")
		}
	}
	setup.EventHub = eventHub

	return nil
}

// InitConfig ...
func (setup *BaseSetupImpl) InitConfig() (apiconfig.Config, error) {
	configImpl, err := config.InitConfig(setup.ConfigFile)
	if err != nil {
		return nil, err
	}
	return configImpl, nil
}

// InstallCC use low level client to install chaincode
func (setup *BaseSetupImpl) InstallCC(name string, path string, version string, ccPackage *fab.CCPackage) error {

	icr := fab.InstallChaincodeRequest{Name: name, Path: path, Version: version, Package: ccPackage, Targets: peer.PeersToTxnProcessors(setup.Channel.Peers())}

	transactionProposalResponse, _, err := setup.Client.InstallChaincode(icr)

	if err != nil {
		return errors.WithMessage(err, "InstallChaincode failed")
	}
	for _, v := range transactionProposalResponse {
		if v.Err != nil {
			return errors.WithMessage(v.Err, "InstallChaincode endorser failed")
		}
	}

	return nil
}

// GetDeployPath ..
func (setup *BaseSetupImpl) GetDeployPath() string {
	pwd, _ := os.Getwd()
	return pwd
}

// InstallAndInstantiateSuningCC install and instantiate using resource management client
func (setup *BaseSetupImpl) InstallAndInstantiateSuningCC() error {

	if setup.ChainCodeID == "" {
		setup.ChainCodeID = GenerateRandomID()
	}

	return setup.InstallAndInstantiateCC(setup.ChainCodeID, "github.com/chaincode", "v1.0", setup.GetDeployPath(), nil)
}

// InstallAndInstantiateCC install and instantiate using resource management client
func (setup *BaseSetupImpl) InstallAndInstantiateCC(ccName, ccPath, ccVersion, goPath string, ccArgs [][]byte) error {

	ccPkg, err := packager.NewCCPackage(ccPath, goPath)
	if err != nil {
		return err
	}

	installCCReq := resmgmt.InstallCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Package: ccPkg}

	_, err = org1ResMgmt.InstallCC(installCCReq)
	if err != nil {
		return err
	}

	_, err = org2ResMgmt.InstallCC(installCCReq)
	if err != nil {
		return err
	}

	// Set the chaincode policy
	ccPolicy := cauthdsl.SignedByAnyMember([]string{"Org1MSP", "Org2MSP"})

	// Org1 resource manager will instantiate cc on suningchannel
	err = org1ResMgmt.InstantiateCC(setup.ChannelID, resmgmt.InstantiateCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Args: ccArgs, Policy: ccPolicy})
	if err != nil {
		return err
	}

	// Org1 resource manager will instantiate cc on suningchannel
	err = org2ResMgmt.InstantiateCC(setup.ChannelID, resmgmt.InstantiateCCRequest{Name: ccName, Path: ccPath, Version: ccVersion, Args: ccArgs, Policy: ccPolicy})
	if err != nil {
		return err
	}

	return nil

}

// GetChannel initializes and returns a channel based on config
func (setup *BaseSetupImpl) GetChannel(client fab.FabricClient, channelID string, orgs []string) (fab.Channel, error) {

	channel, err := client.NewChannel(channelID)
	if err != nil {
		return nil, errors.WithMessage(err, "NewChannel failed")
	}

	ordererConfig, err := client.Config().RandomOrdererConfig()
	if err != nil {
		return nil, errors.WithMessage(err, "RandomOrdererConfig failed")
	}

	orderer, err := orderer.NewOrdererFromConfig(ordererConfig, client.Config())
	if err != nil {
		return nil, errors.WithMessage(err, "NewOrderer failed")
	}
	err = channel.AddOrderer(orderer)
	if err != nil {
		return nil, errors.WithMessage(err, "adding orderer failed")
	}

	for _, org := range orgs {
		peerConfig, err := client.Config().PeersConfig(org)
		if err != nil {
			return nil, errors.WithMessage(err, "reading peer config failed")
		}
		for _, p := range peerConfig {
			endorser, err := deffab.NewPeerFromConfig(&apiconfig.NetworkPeer{PeerConfig: p}, client.Config())
			if err != nil {
				return nil, errors.WithMessage(err, "NewPeer failed")
			}
			err = channel.AddPeer(endorser)
			if err != nil {
				return nil, errors.WithMessage(err, "adding peer failed")
			}
		}
	}

	return channel, nil
}

// CreateAndSendTransactionProposal ... TODO duplicate
func (setup *BaseSetupImpl) CreateAndSendTransactionProposal(channel fab.Channel, chainCodeID string,
	fcn string, args [][]byte, targets []apitxn.ProposalProcessor, transientData map[string][]byte) ([]*apitxn.TransactionProposalResponse, apitxn.TransactionID, error) {

	request := apitxn.ChaincodeInvokeRequest{
		Targets:      targets,
		Fcn:          fcn,
		Args:         args,
		TransientMap: transientData,
		ChaincodeID:  chainCodeID,
	}
	transactionProposalResponses, txnID, err := channel.SendTransactionProposal(request)
	if err != nil {
		return nil, txnID, err
	}

	for _, v := range transactionProposalResponses {
		if v.Err != nil {
			return nil, txnID, errors.Wrapf(v.Err, "endorser %s failed", v.Endorser)
		}
	}

	return transactionProposalResponses, txnID, nil
}

// CreateAndSendTransaction ...
func (setup *BaseSetupImpl) CreateAndSendTransaction(channel fab.Channel, resps []*apitxn.TransactionProposalResponse) (*apitxn.TransactionResponse, error) {

	tx, err := channel.CreateTransaction(resps)
	if err != nil {
		return nil, errors.WithMessage(err, "CreateTransaction failed")
	}

	transactionResponse, err := channel.SendTransaction(tx)
	if err != nil {
		return nil, errors.WithMessage(err, "SendTransaction failed")

	}

	if transactionResponse.Err != nil {
		return nil, errors.Wrapf(transactionResponse.Err, "orderer %s failed", transactionResponse.Orderer)
	}

	return transactionResponse, nil
}

// RegisterTxEvent registers on the given eventhub for the give transaction
// returns a boolean channel which receives true when the event is complete
// and an error channel for errors
// TODO - Duplicate
func (setup *BaseSetupImpl) RegisterTxEvent(txID apitxn.TransactionID, eventHub fab.EventHub) (chan bool, chan error) {
	done := make(chan bool)
	fail := make(chan error)

	eventHub.RegisterTxEvent(txID, func(txId string, errorCode pb.TxValidationCode, err error) {
		if err != nil {
			fmt.Printf("Received error event for txid(%s)", txId)
			fail <- err
		} else {
			fmt.Printf("Received success event for txid(%s)", txId)
			done <- true
		}
	})

	return done, fail
}

// getEventHub initilizes the event hub
func (setup *BaseSetupImpl) getEventHub(client fab.FabricClient) (fab.EventHub, error) {
	eventHub, err := events.NewEventHub(client)
	if err != nil {
		return nil, errors.WithMessage(err, "NewEventHub failed")
	}
	foundEventHub := false
	peerConfig, err := client.Config().PeersConfig(setup.OrgID)
	if err != nil {
		return nil, errors.WithMessage(err, "PeersConfig failed")
	}
	for _, p := range peerConfig {
		if p.URL != "" {
			serverHostOverride := ""
			if str, ok := p.GRPCOptions["ssl-target-name-override"].(string); ok {
				serverHostOverride = str
			}
			eventHub.SetPeerAddr(p.EventURL, p.TLSCACerts.Path, serverHostOverride)
			foundEventHub = true
			break
		}
	}

	if !foundEventHub {
		return nil, errors.New("event hub configuration not found")
	}

	return eventHub, nil
}

func HasPrimaryPeerJoinedChannel(client fab.FabricClient, channel fab.Channel) (bool, error) {
	foundChannel := false
	primaryPeer := channel.PrimaryPeer()
	response, err := client.QueryChannels(primaryPeer)
	if err != nil {
		return false, errors.WithMessage(err, "failed to query channel for primary peer")
	}
	for _, responseChannel := range response.Channels {
		if responseChannel.ChannelId == channel.Name() {
			foundChannel = true
		}
	}

	return foundChannel, nil
}

// GenerateRandomID generates random ID
func GenerateRandomID() string {
	rand.Seed(time.Now().UnixNano())
	return randomString(10)
}

// Utility to create random string of strlen length
func randomString(strlen int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

// Invoke...
func (setup *BaseSetupImpl) Invoke(args [][]byte) (string, error) {

	transientDataMap := make(map[string][]byte)
	transientDataMap["result"] = []byte("Transient data in move funds...")

	transactionProposalResponse, txID, err := setup.CreateAndSendTransactionProposal(setup.Channel, setup.ChainCodeID, "invoke", args, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransactionProposal return error: %v", err)
	}
	// Register for commit event
	done, fail := setup.RegisterTxEvent(txID, setup.EventHub)

	txResponse, err := setup.CreateAndSendTransaction(setup.Channel, transactionProposalResponse)
	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransaction return error: %v", err)
	}
	fmt.Println(txResponse)
	select {
	case <-done:
	case <-fail:
		return "", fmt.Errorf("invoke Error received from eventhub for txid(%s) error(%v)", txID, fail)
	case <-time.After(time.Second * 30):
		return "", fmt.Errorf("invoke Didn't receive block event for txid(%s)", txID)
	}

	return txID.ID, nil
}

func (setup *BaseSetupImpl) Query(args [][]byte) (string, error) {
	transactionProposalResponses, _, err := setup.CreateAndSendTransactionProposal(setup.Channel, setup.ChainCodeID, "invoke", args, []apitxn.ProposalProcessor{setup.Channel.PrimaryPeer()}, nil)
	if err != nil {
		return "", fmt.Errorf("CreateAndSendTransactionProposal return error: %v", err)
	}
	return string(transactionProposalResponses[0].ProposalResponse.GetResponse().Payload), nil
}
