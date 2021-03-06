/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package sdk

import (
	"testing"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func TestResMgmtClientQueries(t *testing.T) {

	// Using shared SDK instance to increase test speed.
	sdk := mainSDK
	testSetup := mainTestSetup
	chaincodeID := mainChaincodeID

	//testSetup := integration.BaseSetupImpl{
	//	ConfigFile:    "../" + integration.ConfigTestFile,
	//	ChannelID:     "mychannel",
	//	OrgID:         org1Name,
	//	ChannelConfig: path.Join("../../../", metadata.ChannelConfigPath, "mychannel.tx"),
	//}

	// Create SDK setup for the integration tests
	//sdk, err := fabsdk.New(config.FromFile(testSetup.ConfigFile))
	//if err != nil {
	//	t.Fatalf("Failed to create new SDK: %s", err)
	//}
	//defer sdk.Close()

	//if err := testSetup.Initialize(sdk); err != nil {
	//	t.Fatalf(err.Error())
	//}

	//ccID := integration.GenerateRandomID()
	//if _, err := integration.InstallAndInstantiateExampleCC(sdk, fabsdk.WithUser("Admin"), testSetup.OrgID, ccID); err != nil {
	//	t.Fatalf("InstallAndInstantiateExampleCC return error: %v", err)
	//}

	//prepare contexts
	org1AdminClientContext := sdk.Context(fabsdk.WithUser(org1AdminUser), fabsdk.WithOrg(org1Name))

	// Resource management client
	client, err := resmgmt.New(org1AdminClientContext)
	if err != nil {
		t.Fatalf("Failed to create new resource management client: %s", err)
	}

	// Our target for queries will be primary peer on this channel
	target := testSetup.Targets[0]

	testQueryConfigFromOrderer(t, testSetup.ChannelID, client)

	testInstalledChaincodes(t, chaincodeID, target, client)

	testInstantiatedChaincodes(t, testSetup.ChannelID, chaincodeID, target, client)

	testQueryChannels(t, testSetup.ChannelID, target, client)

}

func testInstantiatedChaincodes(t *testing.T, channelID string, ccID string, target string, client *resmgmt.Client) {

	chaincodeQueryResponse, err := client.QueryInstantiatedChaincodes(channelID, resmgmt.WithTargetURLs(target))
	if err != nil {
		t.Fatalf("QueryInstantiatedChaincodes return error: %v", err)
	}

	found := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstantiatedCC: %s", chaincode)
		if chaincode.Name == ccID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryInstantiatedChaincodes failed to find instantiated %s chaincode", ccID)
	}
}

func testInstalledChaincodes(t *testing.T, ccID string, target string, client *resmgmt.Client) {

	chaincodeQueryResponse, err := client.QueryInstalledChaincodes(resmgmt.WithTargetURLs(target))
	if err != nil {
		t.Fatalf("QueryInstalledChaincodes return error: %v", err)
	}

	found := false
	for _, chaincode := range chaincodeQueryResponse.Chaincodes {
		t.Logf("**InstalledCC: %s", chaincode)
		if chaincode.Name == ccID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryInstalledChaincodes failed to find installed %s chaincode", ccID)
	}
}

func testQueryChannels(t *testing.T, channelID string, target string, client *resmgmt.Client) {

	channelQueryResponse, err := client.QueryChannels(resmgmt.WithTargetURLs(target))
	if err != nil {
		t.Fatalf("QueryChannels return error: %v", err)
	}

	found := false
	for _, channel := range channelQueryResponse.Channels {
		t.Logf("**Channel: %s", channel)
		if channel.ChannelId == channelID {
			found = true
		}
	}

	if !found {
		t.Fatalf("QueryChannels failed, peer did not join '%s' channel", channelID)
	}

}

func testQueryConfigFromOrderer(t *testing.T, channelID string, client *resmgmt.Client) {

	channelCfg, err := client.QueryConfigFromOrderer(channelID)
	if err != nil {
		t.Fatalf("QueryConfig return error: %v", err)
	}

	expected := "orderer.example.com:7050"
	if !contains(channelCfg.Orderers(), expected) {
		t.Fatalf("Expected orderer %s, got %s", expected, channelCfg.Orderers())
	}

	channelCfg, err = client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererURL("orderer.example.com"))
	if err != nil {
		t.Fatalf("QueryConfig return error: %v", err)
	}
	if !contains(channelCfg.Orderers(), expected) {
		t.Fatalf("Expected orderer %s, got %s", expected, channelCfg.Orderers())
	}

	channelCfg, err = client.QueryConfigFromOrderer(channelID, resmgmt.WithOrdererURL("non-existent"))
	if err == nil {
		t.Fatalf("QueryConfig should have failed for invalid orderer")
	}

}

func contains(list []string, value string) bool {
	for _, e := range list {
		if e == value {
			return true
		}
	}
	return false
}
