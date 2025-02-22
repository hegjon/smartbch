package api

import (
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tendermint/tendermint/libs/log"

	sbchapi "github.com/smartbch/smartbch/api"
	"github.com/smartbch/smartbch/rpc/api/filters"
)

const (
	namespaceEth      = "eth"
	namespaceNet      = "net"
	namespaceWeb3     = "web3"
	namespacePersonal = "personal"
	namespaceEVM      = "evm"
	namespaceSBCH     = "sbch"

	apiVersion = "1.0"
)

// GetAPIs returns the list of all APIs from the Ethereum namespaces
func GetAPIs(backend sbchapi.BackendService,
	logger log.Logger, testKeys []string) []rpc.API {

	_ethAPI := newEthAPI(backend, testKeys, logger)
	_netAPI := newNetAPI(backend.ChainId().Uint64())
	_web3API := web3API{}
	_filterAPI := filters.NewAPI(backend)
	_sbchAPI := newSbchAPI(backend)
	_evmAPI := newEvmAPI(backend)

	return []rpc.API{
		{
			Namespace: namespaceEth,
			Version:   apiVersion,
			Service:   _ethAPI,
			Public:    true,
		},
		{
			Namespace: namespaceEth,
			Version:   apiVersion,
			Service:   _filterAPI,
			Public:    true,
		},
		{
			Namespace: namespaceWeb3,
			Version:   apiVersion,
			Service:   _web3API,
			Public:    true,
		},
		{
			Namespace: namespaceNet,
			Version:   apiVersion,
			Service:   _netAPI,
			Public:    true,
		},
		{
			Namespace: namespaceSBCH,
			Version:   apiVersion,
			Service:   _sbchAPI,
			Public:    true,
		},
		{
			Namespace: namespaceEVM,
			Version:   apiVersion,
			Service:   _evmAPI,
			Public:    true,
		},
	}
}
