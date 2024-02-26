// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package mockbridgex

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// MockBridgeXMetaData contains all meta data concerning the MockBridgeX contract.
var MockBridgeXMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"}],\"name\":\"AuthorizeEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"proofNonce\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"startBlock\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"endBlock\",\"type\":\"uint64\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"dataCommitment\",\"type\":\"bytes32\"}],\"name\":\"DataCommitmentStored\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"_from\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"_to\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"_duration\",\"type\":\"uint256\"}],\"name\":\"SendToSequencerEvent\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"}],\"name\":\"Authorize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"Withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"_targetBlock\",\"type\":\"uint64\"},{\"internalType\":\"bytes32\",\"name\":\"dataCommitment\",\"type\":\"bytes32\"}],\"name\":\"commitHeaderRange\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"string\",\"name\":\"_to\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"_duration\",\"type\":\"uint256\"}],\"name\":\"deposit\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestBlock\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"state_dataCommitments\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"state_proofNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// MockBridgeXABI is the input ABI used to generate the binding from.
// Deprecated: Use MockBridgeXMetaData.ABI instead.
var MockBridgeXABI = MockBridgeXMetaData.ABI

// MockBridgeX is an auto generated Go binding around an Ethereum contract.
type MockBridgeX struct {
	MockBridgeXCaller     // Read-only binding to the contract
	MockBridgeXTransactor // Write-only binding to the contract
	MockBridgeXFilterer   // Log filterer for contract events
}

// MockBridgeXCaller is an auto generated read-only Go binding around an Ethereum contract.
type MockBridgeXCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockBridgeXTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MockBridgeXTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockBridgeXFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MockBridgeXFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockBridgeXSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MockBridgeXSession struct {
	Contract     *MockBridgeX      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MockBridgeXCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MockBridgeXCallerSession struct {
	Contract *MockBridgeXCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// MockBridgeXTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MockBridgeXTransactorSession struct {
	Contract     *MockBridgeXTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// MockBridgeXRaw is an auto generated low-level Go binding around an Ethereum contract.
type MockBridgeXRaw struct {
	Contract *MockBridgeX // Generic contract binding to access the raw methods on
}

// MockBridgeXCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MockBridgeXCallerRaw struct {
	Contract *MockBridgeXCaller // Generic read-only contract binding to access the raw methods on
}

// MockBridgeXTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MockBridgeXTransactorRaw struct {
	Contract *MockBridgeXTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMockBridgeX creates a new instance of MockBridgeX, bound to a specific deployed contract.
func NewMockBridgeX(address common.Address, backend bind.ContractBackend) (*MockBridgeX, error) {
	contract, err := bindMockBridgeX(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MockBridgeX{MockBridgeXCaller: MockBridgeXCaller{contract: contract}, MockBridgeXTransactor: MockBridgeXTransactor{contract: contract}, MockBridgeXFilterer: MockBridgeXFilterer{contract: contract}}, nil
}

// NewMockBridgeXCaller creates a new read-only instance of MockBridgeX, bound to a specific deployed contract.
func NewMockBridgeXCaller(address common.Address, caller bind.ContractCaller) (*MockBridgeXCaller, error) {
	contract, err := bindMockBridgeX(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MockBridgeXCaller{contract: contract}, nil
}

// NewMockBridgeXTransactor creates a new write-only instance of MockBridgeX, bound to a specific deployed contract.
func NewMockBridgeXTransactor(address common.Address, transactor bind.ContractTransactor) (*MockBridgeXTransactor, error) {
	contract, err := bindMockBridgeX(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MockBridgeXTransactor{contract: contract}, nil
}

// NewMockBridgeXFilterer creates a new log filterer instance of MockBridgeX, bound to a specific deployed contract.
func NewMockBridgeXFilterer(address common.Address, filterer bind.ContractFilterer) (*MockBridgeXFilterer, error) {
	contract, err := bindMockBridgeX(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MockBridgeXFilterer{contract: contract}, nil
}

// bindMockBridgeX binds a generic wrapper to an already deployed contract.
func bindMockBridgeX(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MockBridgeXMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockBridgeX *MockBridgeXRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockBridgeX.Contract.MockBridgeXCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockBridgeX *MockBridgeXRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockBridgeX.Contract.MockBridgeXTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockBridgeX *MockBridgeXRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockBridgeX.Contract.MockBridgeXTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockBridgeX *MockBridgeXCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockBridgeX.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockBridgeX *MockBridgeXTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockBridgeX.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockBridgeX *MockBridgeXTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockBridgeX.Contract.contract.Transact(opts, method, params...)
}

// LatestBlock is a free data retrieval call binding the contract method 0x07e2da96.
//
// Solidity: function latestBlock() view returns(uint64)
func (_MockBridgeX *MockBridgeXCaller) LatestBlock(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _MockBridgeX.contract.Call(opts, &out, "latestBlock")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// LatestBlock is a free data retrieval call binding the contract method 0x07e2da96.
//
// Solidity: function latestBlock() view returns(uint64)
func (_MockBridgeX *MockBridgeXSession) LatestBlock() (uint64, error) {
	return _MockBridgeX.Contract.LatestBlock(&_MockBridgeX.CallOpts)
}

// LatestBlock is a free data retrieval call binding the contract method 0x07e2da96.
//
// Solidity: function latestBlock() view returns(uint64)
func (_MockBridgeX *MockBridgeXCallerSession) LatestBlock() (uint64, error) {
	return _MockBridgeX.Contract.LatestBlock(&_MockBridgeX.CallOpts)
}

// StateDataCommitments is a free data retrieval call binding the contract method 0xaeeed33e.
//
// Solidity: function state_dataCommitments(uint256 ) view returns(bytes32)
func (_MockBridgeX *MockBridgeXCaller) StateDataCommitments(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _MockBridgeX.contract.Call(opts, &out, "state_dataCommitments", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// StateDataCommitments is a free data retrieval call binding the contract method 0xaeeed33e.
//
// Solidity: function state_dataCommitments(uint256 ) view returns(bytes32)
func (_MockBridgeX *MockBridgeXSession) StateDataCommitments(arg0 *big.Int) ([32]byte, error) {
	return _MockBridgeX.Contract.StateDataCommitments(&_MockBridgeX.CallOpts, arg0)
}

// StateDataCommitments is a free data retrieval call binding the contract method 0xaeeed33e.
//
// Solidity: function state_dataCommitments(uint256 ) view returns(bytes32)
func (_MockBridgeX *MockBridgeXCallerSession) StateDataCommitments(arg0 *big.Int) ([32]byte, error) {
	return _MockBridgeX.Contract.StateDataCommitments(&_MockBridgeX.CallOpts, arg0)
}

// StateProofNonce is a free data retrieval call binding the contract method 0x55ae3f22.
//
// Solidity: function state_proofNonce() view returns(uint256)
func (_MockBridgeX *MockBridgeXCaller) StateProofNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _MockBridgeX.contract.Call(opts, &out, "state_proofNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StateProofNonce is a free data retrieval call binding the contract method 0x55ae3f22.
//
// Solidity: function state_proofNonce() view returns(uint256)
func (_MockBridgeX *MockBridgeXSession) StateProofNonce() (*big.Int, error) {
	return _MockBridgeX.Contract.StateProofNonce(&_MockBridgeX.CallOpts)
}

// StateProofNonce is a free data retrieval call binding the contract method 0x55ae3f22.
//
// Solidity: function state_proofNonce() view returns(uint256)
func (_MockBridgeX *MockBridgeXCallerSession) StateProofNonce() (*big.Int, error) {
	return _MockBridgeX.Contract.StateProofNonce(&_MockBridgeX.CallOpts)
}

// Authorize is a paid mutator transaction binding the contract method 0xb3907830.
//
// Solidity: function Authorize(bytes _message) returns()
func (_MockBridgeX *MockBridgeXTransactor) Authorize(opts *bind.TransactOpts, _message []byte) (*types.Transaction, error) {
	return _MockBridgeX.contract.Transact(opts, "Authorize", _message)
}

// Authorize is a paid mutator transaction binding the contract method 0xb3907830.
//
// Solidity: function Authorize(bytes _message) returns()
func (_MockBridgeX *MockBridgeXSession) Authorize(_message []byte) (*types.Transaction, error) {
	return _MockBridgeX.Contract.Authorize(&_MockBridgeX.TransactOpts, _message)
}

// Authorize is a paid mutator transaction binding the contract method 0xb3907830.
//
// Solidity: function Authorize(bytes _message) returns()
func (_MockBridgeX *MockBridgeXTransactorSession) Authorize(_message []byte) (*types.Transaction, error) {
	return _MockBridgeX.Contract.Authorize(&_MockBridgeX.TransactOpts, _message)
}

// Withdraw is a paid mutator transaction binding the contract method 0x57ea89b6.
//
// Solidity: function Withdraw() returns()
func (_MockBridgeX *MockBridgeXTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockBridgeX.contract.Transact(opts, "Withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x57ea89b6.
//
// Solidity: function Withdraw() returns()
func (_MockBridgeX *MockBridgeXSession) Withdraw() (*types.Transaction, error) {
	return _MockBridgeX.Contract.Withdraw(&_MockBridgeX.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x57ea89b6.
//
// Solidity: function Withdraw() returns()
func (_MockBridgeX *MockBridgeXTransactorSession) Withdraw() (*types.Transaction, error) {
	return _MockBridgeX.Contract.Withdraw(&_MockBridgeX.TransactOpts)
}

// CommitHeaderRange is a paid mutator transaction binding the contract method 0x6c1f8d13.
//
// Solidity: function commitHeaderRange(uint64 _targetBlock, bytes32 dataCommitment) returns()
func (_MockBridgeX *MockBridgeXTransactor) CommitHeaderRange(opts *bind.TransactOpts, _targetBlock uint64, dataCommitment [32]byte) (*types.Transaction, error) {
	return _MockBridgeX.contract.Transact(opts, "commitHeaderRange", _targetBlock, dataCommitment)
}

// CommitHeaderRange is a paid mutator transaction binding the contract method 0x6c1f8d13.
//
// Solidity: function commitHeaderRange(uint64 _targetBlock, bytes32 dataCommitment) returns()
func (_MockBridgeX *MockBridgeXSession) CommitHeaderRange(_targetBlock uint64, dataCommitment [32]byte) (*types.Transaction, error) {
	return _MockBridgeX.Contract.CommitHeaderRange(&_MockBridgeX.TransactOpts, _targetBlock, dataCommitment)
}

// CommitHeaderRange is a paid mutator transaction binding the contract method 0x6c1f8d13.
//
// Solidity: function commitHeaderRange(uint64 _targetBlock, bytes32 dataCommitment) returns()
func (_MockBridgeX *MockBridgeXTransactorSession) CommitHeaderRange(_targetBlock uint64, dataCommitment [32]byte) (*types.Transaction, error) {
	return _MockBridgeX.Contract.CommitHeaderRange(&_MockBridgeX.TransactOpts, _targetBlock, dataCommitment)
}

// Deposit is a paid mutator transaction binding the contract method 0x89227ec3.
//
// Solidity: function deposit(uint256 _amount, string _to, uint256 _duration) returns()
func (_MockBridgeX *MockBridgeXTransactor) Deposit(opts *bind.TransactOpts, _amount *big.Int, _to string, _duration *big.Int) (*types.Transaction, error) {
	return _MockBridgeX.contract.Transact(opts, "deposit", _amount, _to, _duration)
}

// Deposit is a paid mutator transaction binding the contract method 0x89227ec3.
//
// Solidity: function deposit(uint256 _amount, string _to, uint256 _duration) returns()
func (_MockBridgeX *MockBridgeXSession) Deposit(_amount *big.Int, _to string, _duration *big.Int) (*types.Transaction, error) {
	return _MockBridgeX.Contract.Deposit(&_MockBridgeX.TransactOpts, _amount, _to, _duration)
}

// Deposit is a paid mutator transaction binding the contract method 0x89227ec3.
//
// Solidity: function deposit(uint256 _amount, string _to, uint256 _duration) returns()
func (_MockBridgeX *MockBridgeXTransactorSession) Deposit(_amount *big.Int, _to string, _duration *big.Int) (*types.Transaction, error) {
	return _MockBridgeX.Contract.Deposit(&_MockBridgeX.TransactOpts, _amount, _to, _duration)
}

// MockBridgeXAuthorizeEventIterator is returned from FilterAuthorizeEvent and is used to iterate over the raw logs and unpacked data for AuthorizeEvent events raised by the MockBridgeX contract.
type MockBridgeXAuthorizeEventIterator struct {
	Event *MockBridgeXAuthorizeEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MockBridgeXAuthorizeEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MockBridgeXAuthorizeEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MockBridgeXAuthorizeEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MockBridgeXAuthorizeEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MockBridgeXAuthorizeEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MockBridgeXAuthorizeEvent represents a AuthorizeEvent event raised by the MockBridgeX contract.
type MockBridgeXAuthorizeEvent struct {
	From    common.Address
	Message []byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterAuthorizeEvent is a free log retrieval operation binding the contract event 0x0de3682d77bb5d715a5dba2f9da0d61c2afa6d0e32190e6873a3790e03c5965a.
//
// Solidity: event AuthorizeEvent(address indexed _from, bytes _message)
func (_MockBridgeX *MockBridgeXFilterer) FilterAuthorizeEvent(opts *bind.FilterOpts, _from []common.Address) (*MockBridgeXAuthorizeEventIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _MockBridgeX.contract.FilterLogs(opts, "AuthorizeEvent", _fromRule)
	if err != nil {
		return nil, err
	}
	return &MockBridgeXAuthorizeEventIterator{contract: _MockBridgeX.contract, event: "AuthorizeEvent", logs: logs, sub: sub}, nil
}

// WatchAuthorizeEvent is a free log subscription operation binding the contract event 0x0de3682d77bb5d715a5dba2f9da0d61c2afa6d0e32190e6873a3790e03c5965a.
//
// Solidity: event AuthorizeEvent(address indexed _from, bytes _message)
func (_MockBridgeX *MockBridgeXFilterer) WatchAuthorizeEvent(opts *bind.WatchOpts, sink chan<- *MockBridgeXAuthorizeEvent, _from []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _MockBridgeX.contract.WatchLogs(opts, "AuthorizeEvent", _fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MockBridgeXAuthorizeEvent)
				if err := _MockBridgeX.contract.UnpackLog(event, "AuthorizeEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAuthorizeEvent is a log parse operation binding the contract event 0x0de3682d77bb5d715a5dba2f9da0d61c2afa6d0e32190e6873a3790e03c5965a.
//
// Solidity: event AuthorizeEvent(address indexed _from, bytes _message)
func (_MockBridgeX *MockBridgeXFilterer) ParseAuthorizeEvent(log types.Log) (*MockBridgeXAuthorizeEvent, error) {
	event := new(MockBridgeXAuthorizeEvent)
	if err := _MockBridgeX.contract.UnpackLog(event, "AuthorizeEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MockBridgeXDataCommitmentStoredIterator is returned from FilterDataCommitmentStored and is used to iterate over the raw logs and unpacked data for DataCommitmentStored events raised by the MockBridgeX contract.
type MockBridgeXDataCommitmentStoredIterator struct {
	Event *MockBridgeXDataCommitmentStored // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MockBridgeXDataCommitmentStoredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MockBridgeXDataCommitmentStored)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MockBridgeXDataCommitmentStored)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MockBridgeXDataCommitmentStoredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MockBridgeXDataCommitmentStoredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MockBridgeXDataCommitmentStored represents a DataCommitmentStored event raised by the MockBridgeX contract.
type MockBridgeXDataCommitmentStored struct {
	ProofNonce     *big.Int
	StartBlock     uint64
	EndBlock       uint64
	DataCommitment [32]byte
	Raw            types.Log // Blockchain specific contextual infos
}

// FilterDataCommitmentStored is a free log retrieval operation binding the contract event 0x34dd3689f5bd77a60a3ff2e09483dcab032fa2f1fd7227af3e24bed21beab1cb.
//
// Solidity: event DataCommitmentStored(uint256 proofNonce, uint64 indexed startBlock, uint64 indexed endBlock, bytes32 indexed dataCommitment)
func (_MockBridgeX *MockBridgeXFilterer) FilterDataCommitmentStored(opts *bind.FilterOpts, startBlock []uint64, endBlock []uint64, dataCommitment [][32]byte) (*MockBridgeXDataCommitmentStoredIterator, error) {

	var startBlockRule []interface{}
	for _, startBlockItem := range startBlock {
		startBlockRule = append(startBlockRule, startBlockItem)
	}
	var endBlockRule []interface{}
	for _, endBlockItem := range endBlock {
		endBlockRule = append(endBlockRule, endBlockItem)
	}
	var dataCommitmentRule []interface{}
	for _, dataCommitmentItem := range dataCommitment {
		dataCommitmentRule = append(dataCommitmentRule, dataCommitmentItem)
	}

	logs, sub, err := _MockBridgeX.contract.FilterLogs(opts, "DataCommitmentStored", startBlockRule, endBlockRule, dataCommitmentRule)
	if err != nil {
		return nil, err
	}
	return &MockBridgeXDataCommitmentStoredIterator{contract: _MockBridgeX.contract, event: "DataCommitmentStored", logs: logs, sub: sub}, nil
}

// WatchDataCommitmentStored is a free log subscription operation binding the contract event 0x34dd3689f5bd77a60a3ff2e09483dcab032fa2f1fd7227af3e24bed21beab1cb.
//
// Solidity: event DataCommitmentStored(uint256 proofNonce, uint64 indexed startBlock, uint64 indexed endBlock, bytes32 indexed dataCommitment)
func (_MockBridgeX *MockBridgeXFilterer) WatchDataCommitmentStored(opts *bind.WatchOpts, sink chan<- *MockBridgeXDataCommitmentStored, startBlock []uint64, endBlock []uint64, dataCommitment [][32]byte) (event.Subscription, error) {

	var startBlockRule []interface{}
	for _, startBlockItem := range startBlock {
		startBlockRule = append(startBlockRule, startBlockItem)
	}
	var endBlockRule []interface{}
	for _, endBlockItem := range endBlock {
		endBlockRule = append(endBlockRule, endBlockItem)
	}
	var dataCommitmentRule []interface{}
	for _, dataCommitmentItem := range dataCommitment {
		dataCommitmentRule = append(dataCommitmentRule, dataCommitmentItem)
	}

	logs, sub, err := _MockBridgeX.contract.WatchLogs(opts, "DataCommitmentStored", startBlockRule, endBlockRule, dataCommitmentRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MockBridgeXDataCommitmentStored)
				if err := _MockBridgeX.contract.UnpackLog(event, "DataCommitmentStored", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDataCommitmentStored is a log parse operation binding the contract event 0x34dd3689f5bd77a60a3ff2e09483dcab032fa2f1fd7227af3e24bed21beab1cb.
//
// Solidity: event DataCommitmentStored(uint256 proofNonce, uint64 indexed startBlock, uint64 indexed endBlock, bytes32 indexed dataCommitment)
func (_MockBridgeX *MockBridgeXFilterer) ParseDataCommitmentStored(log types.Log) (*MockBridgeXDataCommitmentStored, error) {
	event := new(MockBridgeXDataCommitmentStored)
	if err := _MockBridgeX.contract.UnpackLog(event, "DataCommitmentStored", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MockBridgeXSendToSequencerEventIterator is returned from FilterSendToSequencerEvent and is used to iterate over the raw logs and unpacked data for SendToSequencerEvent events raised by the MockBridgeX contract.
type MockBridgeXSendToSequencerEventIterator struct {
	Event *MockBridgeXSendToSequencerEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MockBridgeXSendToSequencerEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MockBridgeXSendToSequencerEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MockBridgeXSendToSequencerEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MockBridgeXSendToSequencerEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MockBridgeXSendToSequencerEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MockBridgeXSendToSequencerEvent represents a SendToSequencerEvent event raised by the MockBridgeX contract.
type MockBridgeXSendToSequencerEvent struct {
	From     common.Address
	Amount   *big.Int
	To       string
	Duration *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSendToSequencerEvent is a free log retrieval operation binding the contract event 0x5dee65305d37f37b03a10fb088c878b533e94440a61a4ad4f99cb82821398f98.
//
// Solidity: event SendToSequencerEvent(address indexed _from, uint256 _amount, string _to, uint256 _duration)
func (_MockBridgeX *MockBridgeXFilterer) FilterSendToSequencerEvent(opts *bind.FilterOpts, _from []common.Address) (*MockBridgeXSendToSequencerEventIterator, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _MockBridgeX.contract.FilterLogs(opts, "SendToSequencerEvent", _fromRule)
	if err != nil {
		return nil, err
	}
	return &MockBridgeXSendToSequencerEventIterator{contract: _MockBridgeX.contract, event: "SendToSequencerEvent", logs: logs, sub: sub}, nil
}

// WatchSendToSequencerEvent is a free log subscription operation binding the contract event 0x5dee65305d37f37b03a10fb088c878b533e94440a61a4ad4f99cb82821398f98.
//
// Solidity: event SendToSequencerEvent(address indexed _from, uint256 _amount, string _to, uint256 _duration)
func (_MockBridgeX *MockBridgeXFilterer) WatchSendToSequencerEvent(opts *bind.WatchOpts, sink chan<- *MockBridgeXSendToSequencerEvent, _from []common.Address) (event.Subscription, error) {

	var _fromRule []interface{}
	for _, _fromItem := range _from {
		_fromRule = append(_fromRule, _fromItem)
	}

	logs, sub, err := _MockBridgeX.contract.WatchLogs(opts, "SendToSequencerEvent", _fromRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MockBridgeXSendToSequencerEvent)
				if err := _MockBridgeX.contract.UnpackLog(event, "SendToSequencerEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSendToSequencerEvent is a log parse operation binding the contract event 0x5dee65305d37f37b03a10fb088c878b533e94440a61a4ad4f99cb82821398f98.
//
// Solidity: event SendToSequencerEvent(address indexed _from, uint256 _amount, string _to, uint256 _duration)
func (_MockBridgeX *MockBridgeXFilterer) ParseSendToSequencerEvent(log types.Log) (*MockBridgeXSendToSequencerEvent, error) {
	event := new(MockBridgeXSendToSequencerEvent)
	if err := _MockBridgeX.contract.UnpackLog(event, "SendToSequencerEvent", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
