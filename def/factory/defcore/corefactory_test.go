/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package defcore

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig"
	"github.com/hyperledger/fabric-sdk-go/api/apiconfig/mocks"
	"github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/def/fabapi/opt"
	"github.com/hyperledger/fabric-sdk-go/def/provider/fabpvdr"
	configImpl "github.com/hyperledger/fabric-sdk-go/pkg/config"
	cryptosuitewrapper "github.com/hyperledger/fabric-sdk-go/pkg/cryptosuite/bccsp/wrapper"
	kvs "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/keyvaluestore"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	signingMgr "github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/signingmgr"
	"github.com/hyperledger/fabric-sdk-go/pkg/logging/deflogger"
)

func TestNewConfigProvider(t *testing.T) {
	factory := NewProviderFactory()

	configOpts := opt.ConfigOpts{}
	sdkOpts := opt.SDKOpts{}

	config, err := factory.NewConfigProvider(configOpts, sdkOpts)
	if err != nil {
		t.Fatalf("Unexpected error creating config provider %v", err)
	}

	_, ok := config.(*configImpl.Config)
	if !ok {
		t.Fatalf("Unexpected config provider created")
	}
}

func TestNewStateStoreProvider(t *testing.T) {
	factory := NewProviderFactory()

	opts := opt.StateStoreOpts{
		Path: "/tmp/fabsdkgo_test/store",
	}
	config := mocks.NewMockConfig()

	stateStore, err := factory.NewStateStoreProvider(opts, config)
	if err != nil {
		t.Fatalf("Unexpected error creating state store provider %v", err)
	}

	_, ok := stateStore.(*kvs.FileKeyValueStore)
	if !ok {
		t.Fatalf("Unexpected state store provider created")
	}
}

func newMockStateStore(t *testing.T) apifabclient.KeyValueStore {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockClientConfig := apiconfig.ClientConfig{
		CredentialStore: apiconfig.CredentialStoreType{
			Path: "/tmp/fabsdkgo_test/store",
		},
	}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)

	opts := opt.StateStoreOpts{}
	stateStore, err := factory.NewStateStoreProvider(opts, mockConfig)
	if err != nil {
		t.Fatalf("Unexpected error creating state store provider %v", err)
	}
	return stateStore
}
func TestNewStateStoreProviderByConfig(t *testing.T) {
	stateStore := newMockStateStore(t)

	_, ok := stateStore.(*kvs.FileKeyValueStore)
	if !ok {
		t.Fatalf("Unexpected state store provider created")
	}
}

func TestNewStateStoreProviderEmptyConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockClientConfig := apiconfig.ClientConfig{}
	mockConfig.EXPECT().Client().Return(&mockClientConfig, nil)
	opts := opt.StateStoreOpts{}

	_, err := factory.NewStateStoreProvider(opts, mockConfig)
	if err == nil {
		t.Fatal("Expected error creating state store provider")
	}
}

func TestNewStateStoreProviderFailConfig(t *testing.T) {
	factory := NewProviderFactory()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockConfig := mock_apiconfig.NewMockConfig(mockCtrl)

	mockConfig.EXPECT().Client().Return(nil, errors.New("error"))
	opts := opt.StateStoreOpts{}

	_, err := factory.NewStateStoreProvider(opts, mockConfig)
	if err == nil {
		t.Fatal("Expected error creating state store provider")
	}
}

func TestNewCryptoSuiteProvider(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	cryptosuite, err := factory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	_, ok := cryptosuite.(*cryptosuitewrapper.CryptoSuite)
	if !ok {
		t.Fatalf("Unexpected cryptosuite provider created")
	}
}

func TestNewSigningManager(t *testing.T) {
	factory := NewProviderFactory()
	config := mocks.NewMockConfig()

	cryptosuite, err := factory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := factory.NewSigningManager(cryptosuite, config)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	_, ok := signer.(*signingMgr.SigningManager)
	if !ok {
		t.Fatalf("Unexpected signing manager created")
	}
}

func TestNewFactoryFabricProvider(t *testing.T) {
	factory := NewProviderFactory()

	config := mocks.NewMockConfig()

	cryptosuite, err := factory.NewCryptoSuiteProvider(config)
	if err != nil {
		t.Fatalf("Unexpected error creating cryptosuite provider %v", err)
	}

	signer, err := factory.NewSigningManager(cryptosuite, config)
	if err != nil {
		t.Fatalf("Unexpected error creating signing manager %v", err)
	}

	stateStore := newMockStateStore(t)

	fabricProvider, err := factory.NewFabricProvider(config, stateStore, cryptosuite, signer)
	if err != nil {
		t.Fatalf("Unexpected error creating fabric provider %v", err)
	}

	_, ok := fabricProvider.(*fabpvdr.FabricProvider)
	if !ok {
		t.Fatalf("Unexpected fabric provider created")
	}
}

func TestNewLoggingProvider(t *testing.T) {
	logger := NewLoggerProvider()

	_, ok := logger.(*deflogger.Provider)
	if !ok {
		t.Fatalf("Unexpected logger provider created")
	}
}