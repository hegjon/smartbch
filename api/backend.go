package api

import (
	"context"
	"errors"
	"math/big"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/mempool"
	"github.com/tendermint/tendermint/node"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/ethereum/go-ethereum/common"
	gethcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/bloombits"
	gethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/smartbch/moeingevm/types"
	"github.com/smartbch/smartbch/app"
)

var _ BackendService = &apiBackend{}

const (
	// Ethereum Wire Protocol
	// https://github.com/ethereum/devp2p/blob/master/caps/eth.md
	protocolVersion = 63
)

type apiBackend struct {
	//extRPCEnabled bool
	node *node.Node
	app  *app.App
	//gpo *gasprice.Oracle

	chainSideFeed event.Feed
	chainHeadFeed event.Feed
	blockProcFeed event.Feed
	txFeed        event.Feed
	logsFeed      event.Feed
	rmLogsFeed    event.Feed
	//pendingLogsFeed event.Feed
}

func NewBackend(node *node.Node, app *app.App) BackendService {
	return &apiBackend{
		node: node,
		app:  app,
	}
}

func (backend *apiBackend) ChainId() *big.Int {
	return backend.app.ChainID().ToBig()
}

func (backend *apiBackend) GetStorageAt(address common.Address, key string, blockNumber int64) []byte {
	if blockNumber != int64(rpc.LatestBlockNumber) {
		// TODO: not supported yet
		return nil
	}

	ctx := backend.app.GetContext(app.RpcMode)
	defer ctx.Close(false)

	acc := ctx.GetAccount(address)
	if acc == nil {
		return nil
	}
	return ctx.GetStorageAt(acc.Sequence(), key)
}

func (backend *apiBackend) GetCode(contract common.Address, blockNumber int64) (bytecode []byte, codeHash []byte) {
	if blockNumber != int64(rpc.LatestBlockNumber) {
		// TODO: not supported yet
		return
	}

	ctx := backend.app.GetContext(app.RpcMode)
	defer ctx.Close(false)

	info := ctx.GetCode(contract)
	if info != nil {
		bytecode = info.BytecodeSlice()
		codeHash = info.CodeHashSlice()
	}
	return
}

func (backend *apiBackend) GetBalance(owner common.Address, height int64) (*big.Int, error) {
	ctx := backend.app.GetContext(app.RpcMode)
	defer ctx.Close(false)
	b, err := ctx.GetBalance(owner, height)
	if err != nil {
		return nil, err
	}
	return b.ToBig(), nil
}

func (backend *apiBackend) GetNonce(address common.Address) (uint64, error) {
	ctx := backend.app.GetContext(app.RpcMode)
	defer ctx.Close(false)
	if acc := ctx.GetAccount(address); acc != nil {
		return acc.Nonce(), nil
	}

	return 0, types.ErrAccNotFound
}

func (backend *apiBackend) GetTransaction(txHash common.Hash) (tx *types.Transaction, blockHash common.Hash, blockNumber uint64, blockIndex uint64, err error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)

	if tx, err = ctx.GetTxByHash(txHash); err != nil {
		return
	}
	if tx != nil {
		blockHash = tx.BlockHash
		blockNumber = uint64(tx.BlockNumber)
		blockIndex = uint64(tx.TransactionIndex)
	} else {
		err = errors.New("tx with specific hash not exist")
	}
	return
}

func (backend *apiBackend) BlockByHash(hash common.Hash) (*types.Block, error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)
	block, err := ctx.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (backend *apiBackend) BlockByNumber(number int64) (*types.Block, error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)
	return ctx.GetBlockByHeight(uint64(number))
}

func (backend *apiBackend) ProtocolVersion() int {
	return protocolVersion
}

func (backend *apiBackend) LatestHeight() int64 {
	appCtx := backend.app.GetContext(app.HistoryOnlyMode)
	return appCtx.GetLatestHeight()
}

func (backend *apiBackend) CurrentBlock() (*types.Block, error) {
	appCtx := backend.app.GetContext(app.HistoryOnlyMode)
	block, err := appCtx.GetBlockByHeight(uint64(appCtx.GetLatestHeight()))
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (backend *apiBackend) SendRawTx(signedTx []byte) (common.Hash, error) {
	return backend.broadcastTxSync(signedTx)
}

func (backend *apiBackend) broadcastTxSync(tx tmtypes.Tx) (common.Hash, error) {
	resCh := make(chan *abci.Response, 1)
	err := backend.node.Mempool().CheckTx(tx, func(res *abci.Response) {
		resCh <- res
	}, mempool.TxInfo{})
	if err != nil {
		return common.Hash{}, err
	}
	res := <-resCh
	r := res.GetCheckTx()
	if r.Code != abci.CodeTypeOK {
		return common.Hash{}, errors.New(r.String())
	}
	return common.BytesToHash(tx.Hash()), nil
}

func (backend *apiBackend) Call(tx *gethtypes.Transaction, sender common.Address) (statusCode int, retData []byte) {
	runner, _ := backend.app.RunTxForRpc(tx, sender, false)
	return runner.Status, runner.OutData
}

func (backend *apiBackend) EstimateGas(tx *gethtypes.Transaction, sender common.Address) (statusCode int, retData []byte, gas int64) {
	runner, gas := backend.app.RunTxForRpc(tx, sender, true)
	return runner.Status, runner.OutData, gas
}

func (backend *apiBackend) QueryLogs(addresses []common.Address, topics [][]common.Hash, startHeight, endHeight uint32) ([]types.Log, error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)

	return ctx.QueryLogs(addresses, topics, startHeight, endHeight)
}

func (backend *apiBackend) QueryTxBySrc(addr common.Address, startHeight, endHeight uint32) (tx []*types.Transaction, err error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)
	return ctx.QueryTxBySrc(addr, startHeight, endHeight)
}

func (backend *apiBackend) QueryTxByDst(addr common.Address, startHeight, endHeight uint32) (tx []*types.Transaction, err error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)
	return ctx.QueryTxByDst(addr, startHeight, endHeight)
}

func (backend *apiBackend) QueryTxByAddr(addr common.Address, startHeight, endHeight uint32) (tx []*types.Transaction, err error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)
	return ctx.QueryTxByAddr(addr, startHeight, endHeight)
}

func (backend *apiBackend) MoeQueryLogs(addr common.Address, topics []common.Hash, startHeight, endHeight uint32) ([]types.Log, error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)

	return ctx.BasicQueryLogs(addr, topics, startHeight, endHeight)
}

func (backend *apiBackend) GetTxListByHeight(height uint32) (tx []*types.Transaction, err error) {
	ctx := backend.app.GetContext(app.HistoryOnlyMode)
	defer ctx.Close(false)

	return ctx.GetTxListByHeight(height)
}

func (backend *apiBackend) HeaderByNumber(ctx context.Context, blockNr rpc.BlockNumber) (*types.Header, error) {
	if blockNr == rpc.LatestBlockNumber {
		blockNr = rpc.BlockNumber(backend.app.GetLatestBlockNum())
	}

	appCtx := backend.app.GetContext(app.HistoryOnlyMode)
	defer appCtx.Close(false)
	block, err := appCtx.GetBlockByHeight(uint64(blockNr))
	if err != nil {
		return nil, nil
	}
	return &types.Header{
		Number:    uint64(block.Number),
		BlockHash: block.Hash,
		Bloom:     block.LogsBloom,
	}, nil
}
func (backend *apiBackend) HeaderByHash(ctx context.Context, blockHash common.Hash) (*types.Header, error) {
	appCtx := backend.app.GetContext(app.HistoryOnlyMode)
	defer appCtx.Close(false)
	block, err := appCtx.GetBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}
	return &types.Header{
		Number:    uint64(block.Number),
		BlockHash: block.Hash,
	}, nil
}
func (backend *apiBackend) GetReceipts(ctx context.Context, blockHash common.Hash) (gethtypes.Receipts, error) {
	appCtx := backend.app.GetContext(app.HistoryOnlyMode)
	defer appCtx.Close(false)

	receipts := make([]*gethtypes.Receipt, 0, 8)

	// TODO: query receipts
	//block, err := appCtx.GetBlockByHash(blockHash)
	//if err == nil && block != nil {
	//	for _, txHash := range block.Transactions {
	//		tx, err := appCtx.GetTxByHash(txHash)
	//		if err == nil && tx != nil {
	//			receipts = append(receipts, toGethReceipt(tx))
	//		}
	//	}
	//}
	return receipts, nil
}

func (backend *apiBackend) GetLogs(ctx context.Context, blockHash common.Hash) ([][]*gethtypes.Log, error) {
	appCtx := backend.app.GetContext(app.HistoryOnlyMode)
	defer appCtx.Close(false)

	logs := make([][]*gethtypes.Log, 0, 8)

	block, err := appCtx.GetBlockByHash(blockHash)
	if err == nil && block != nil {
		for _, txHash := range block.Transactions {
			tx, err := appCtx.GetTxByHash(txHash)
			if err == nil && tx != nil {
				txLogs := types.ToGethLogs(tx.Logs)
				// fix log.TxHash
				for _, txLog := range txLogs {
					txLog.TxHash = tx.Hash
				}
				logs = append(logs, txLogs)
			}
		}
	}

	return logs, nil
}

func (backend *apiBackend) SubscribeChainEvent(ch chan<- types.ChainEvent) event.Subscription {
	return backend.app.SubscribeChainEvent(ch)
}
func (backend *apiBackend) SubscribeLogsEvent(ch chan<- []*gethtypes.Log) event.Subscription {
	return backend.app.SubscribeLogsEvent(ch)
}
func (backend *apiBackend) SubscribeNewTxsEvent(ch chan<- gethcore.NewTxsEvent) event.Subscription {
	return backend.txFeed.Subscribe(ch)
}
func (backend *apiBackend) SubscribeRemovedLogsEvent(ch chan<- gethcore.RemovedLogsEvent) event.Subscription {
	return backend.rmLogsFeed.Subscribe(ch)
}

//func (b2 *apiBackend) SubscribePendingLogsEvent(ch chan<- []*gethtypes.Log) event.Subscription {
//	return b2.pendingLogsFeed.Subscribe(ch)
//}

func (backend *apiBackend) BloomStatus() (uint64, uint64) {
	return 4096, 0 // TODO: this is temporary implementation
}
func (backend *apiBackend) ServiceFilter(ctx context.Context, session *bloombits.MatcherSession) {
	panic("implement me")
}
