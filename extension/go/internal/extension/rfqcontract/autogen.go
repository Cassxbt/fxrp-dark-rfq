// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package rfqcontract

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

// RfqSettlementFill is an auto generated low-level Go binding around an user-defined struct.
type RfqSettlementFill struct {
	RfqId  [32]byte
	Taker  common.Address
	Maker  common.Address
	Side   uint8
	Size   *big.Int
	Price  *big.Int
	Expiry *big.Int
}

// RfqSettlementMetaData contains all meta data concerning the RfqSettlement contract.
var RfqSettlementMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_baseToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"_quoteToken\",\"type\":\"address\",\"internalType\":\"contractIERC20\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"baseToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip712Domain\",\"inputs\":[],\"outputs\":[{\"name\":\"fields\",\"type\":\"bytes1\",\"internalType\":\"bytes1\"},{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"version\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"chainId\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"verifyingContract\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"salt\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"extensions\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ftso\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIFtsoV2\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ftsoFeedId\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes21\",\"internalType\":\"bytes21\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ftsoMaxStaleness\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ftsoToleranceBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"hashFill\",\"inputs\":[{\"name\":\"fill\",\"type\":\"tuple\",\"internalType\":\"structRfqSettlement.Fill\",\"components\":[{\"name\":\"rfqId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"taker\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"maker\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumRfqSettlement.Side\"},{\"name\":\"size\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isAttestedSigner\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"quoteToken\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractIERC20\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setAttestedSigner\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"allowed\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFtsoBound\",\"inputs\":[{\"name\":\"_ftso\",\"type\":\"address\",\"internalType\":\"contractIFtsoV2\"},{\"name\":\"_feedId\",\"type\":\"bytes21\",\"internalType\":\"bytes21\"},{\"name\":\"_toleranceBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxStaleness\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"settle\",\"inputs\":[{\"name\":\"fill\",\"type\":\"tuple\",\"internalType\":\"structRfqSettlement.Fill\",\"components\":[{\"name\":\"rfqId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"taker\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"maker\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"side\",\"type\":\"uint8\",\"internalType\":\"enumRfqSettlement.Side\"},{\"name\":\"size\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"attestationSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"settled\",\"inputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"AttestedSignerSet\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"allowed\",\"type\":\"bool\",\"indexed\":false,\"internalType\":\"bool\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EIP712DomainChanged\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Filled\",\"inputs\":[{\"name\":\"rfqId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"taker\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"maker\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"side\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"enumRfqSettlement.Side\"},{\"name\":\"size\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"price\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FtsoBoundSet\",\"inputs\":[{\"name\":\"ftso\",\"type\":\"address\",\"indexed\":false,\"internalType\":\"address\"},{\"name\":\"feedId\",\"type\":\"bytes21\",\"indexed\":false,\"internalType\":\"bytes21\"},{\"name\":\"toleranceBps\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"},{\"name\":\"maxStaleness\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AlreadySettled\",\"inputs\":[{\"name\":\"rfqId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"Expired\",\"inputs\":[{\"name\":\"expiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nowTs\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"FeedDecimalsOutOfRange\",\"inputs\":[{\"name\":\"refDecimals\",\"type\":\"int8\",\"internalType\":\"int8\"}]},{\"type\":\"error\",\"name\":\"InvalidShortString\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidToleranceBps\",\"inputs\":[{\"name\":\"toleranceBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"OwnableInvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"OwnableUnauthorizedAccount\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"PriceOutOfBounds\",\"inputs\":[{\"name\":\"price\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"refPrice\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"toleranceBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"StaleFeed\",\"inputs\":[{\"name\":\"feedTimestamp\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"nowTs\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxStaleness\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"StringTooLong\",\"inputs\":[{\"name\":\"str\",\"type\":\"string\",\"internalType\":\"string\"}]},{\"type\":\"error\",\"name\":\"UntrustedSigner\",\"inputs\":[{\"name\":\"signer\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ZeroAmount\",\"inputs\":[{\"name\":\"size\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"quoteAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
}

// RfqSettlementABI is the input ABI used to generate the binding from.
// Deprecated: Use RfqSettlementMetaData.ABI instead.
var RfqSettlementABI = RfqSettlementMetaData.ABI

// RfqSettlement is an auto generated Go binding around an Ethereum contract.
type RfqSettlement struct {
	RfqSettlementCaller     // Read-only binding to the contract
	RfqSettlementTransactor // Write-only binding to the contract
	RfqSettlementFilterer   // Log filterer for contract events
}

// RfqSettlementCaller is an auto generated read-only Go binding around an Ethereum contract.
type RfqSettlementCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RfqSettlementTransactor is an auto generated write-only Go binding around an Ethereum contract.
type RfqSettlementTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RfqSettlementFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type RfqSettlementFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// RfqSettlementSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type RfqSettlementSession struct {
	Contract     *RfqSettlement    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// RfqSettlementCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type RfqSettlementCallerSession struct {
	Contract *RfqSettlementCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// RfqSettlementTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type RfqSettlementTransactorSession struct {
	Contract     *RfqSettlementTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// RfqSettlementRaw is an auto generated low-level Go binding around an Ethereum contract.
type RfqSettlementRaw struct {
	Contract *RfqSettlement // Generic contract binding to access the raw methods on
}

// RfqSettlementCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type RfqSettlementCallerRaw struct {
	Contract *RfqSettlementCaller // Generic read-only contract binding to access the raw methods on
}

// RfqSettlementTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type RfqSettlementTransactorRaw struct {
	Contract *RfqSettlementTransactor // Generic write-only contract binding to access the raw methods on
}

// NewRfqSettlement creates a new instance of RfqSettlement, bound to a specific deployed contract.
func NewRfqSettlement(address common.Address, backend bind.ContractBackend) (*RfqSettlement, error) {
	contract, err := bindRfqSettlement(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &RfqSettlement{RfqSettlementCaller: RfqSettlementCaller{contract: contract}, RfqSettlementTransactor: RfqSettlementTransactor{contract: contract}, RfqSettlementFilterer: RfqSettlementFilterer{contract: contract}}, nil
}

// NewRfqSettlementCaller creates a new read-only instance of RfqSettlement, bound to a specific deployed contract.
func NewRfqSettlementCaller(address common.Address, caller bind.ContractCaller) (*RfqSettlementCaller, error) {
	contract, err := bindRfqSettlement(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &RfqSettlementCaller{contract: contract}, nil
}

// NewRfqSettlementTransactor creates a new write-only instance of RfqSettlement, bound to a specific deployed contract.
func NewRfqSettlementTransactor(address common.Address, transactor bind.ContractTransactor) (*RfqSettlementTransactor, error) {
	contract, err := bindRfqSettlement(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &RfqSettlementTransactor{contract: contract}, nil
}

// NewRfqSettlementFilterer creates a new log filterer instance of RfqSettlement, bound to a specific deployed contract.
func NewRfqSettlementFilterer(address common.Address, filterer bind.ContractFilterer) (*RfqSettlementFilterer, error) {
	contract, err := bindRfqSettlement(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &RfqSettlementFilterer{contract: contract}, nil
}

// bindRfqSettlement binds a generic wrapper to an already deployed contract.
func bindRfqSettlement(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := RfqSettlementMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RfqSettlement *RfqSettlementRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RfqSettlement.Contract.RfqSettlementCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RfqSettlement *RfqSettlementRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RfqSettlement.Contract.RfqSettlementTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RfqSettlement *RfqSettlementRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RfqSettlement.Contract.RfqSettlementTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_RfqSettlement *RfqSettlementCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _RfqSettlement.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_RfqSettlement *RfqSettlementTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RfqSettlement.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_RfqSettlement *RfqSettlementTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _RfqSettlement.Contract.contract.Transact(opts, method, params...)
}

// BaseToken is a free data retrieval call binding the contract method 0xc55dae63.
//
// Solidity: function baseToken() view returns(address)
func (_RfqSettlement *RfqSettlementCaller) BaseToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "baseToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BaseToken is a free data retrieval call binding the contract method 0xc55dae63.
//
// Solidity: function baseToken() view returns(address)
func (_RfqSettlement *RfqSettlementSession) BaseToken() (common.Address, error) {
	return _RfqSettlement.Contract.BaseToken(&_RfqSettlement.CallOpts)
}

// BaseToken is a free data retrieval call binding the contract method 0xc55dae63.
//
// Solidity: function baseToken() view returns(address)
func (_RfqSettlement *RfqSettlementCallerSession) BaseToken() (common.Address, error) {
	return _RfqSettlement.Contract.BaseToken(&_RfqSettlement.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_RfqSettlement *RfqSettlementCaller) Eip712Domain(opts *bind.CallOpts) (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "eip712Domain")

	outstruct := new(struct {
		Fields            [1]byte
		Name              string
		Version           string
		ChainId           *big.Int
		VerifyingContract common.Address
		Salt              [32]byte
		Extensions        []*big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Fields = *abi.ConvertType(out[0], new([1]byte)).(*[1]byte)
	outstruct.Name = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Version = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.ChainId = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.VerifyingContract = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Salt = *abi.ConvertType(out[5], new([32]byte)).(*[32]byte)
	outstruct.Extensions = *abi.ConvertType(out[6], new([]*big.Int)).(*[]*big.Int)

	return *outstruct, err

}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_RfqSettlement *RfqSettlementSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _RfqSettlement.Contract.Eip712Domain(&_RfqSettlement.CallOpts)
}

// Eip712Domain is a free data retrieval call binding the contract method 0x84b0196e.
//
// Solidity: function eip712Domain() view returns(bytes1 fields, string name, string version, uint256 chainId, address verifyingContract, bytes32 salt, uint256[] extensions)
func (_RfqSettlement *RfqSettlementCallerSession) Eip712Domain() (struct {
	Fields            [1]byte
	Name              string
	Version           string
	ChainId           *big.Int
	VerifyingContract common.Address
	Salt              [32]byte
	Extensions        []*big.Int
}, error) {
	return _RfqSettlement.Contract.Eip712Domain(&_RfqSettlement.CallOpts)
}

// Ftso is a free data retrieval call binding the contract method 0x0da0c022.
//
// Solidity: function ftso() view returns(address)
func (_RfqSettlement *RfqSettlementCaller) Ftso(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "ftso")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Ftso is a free data retrieval call binding the contract method 0x0da0c022.
//
// Solidity: function ftso() view returns(address)
func (_RfqSettlement *RfqSettlementSession) Ftso() (common.Address, error) {
	return _RfqSettlement.Contract.Ftso(&_RfqSettlement.CallOpts)
}

// Ftso is a free data retrieval call binding the contract method 0x0da0c022.
//
// Solidity: function ftso() view returns(address)
func (_RfqSettlement *RfqSettlementCallerSession) Ftso() (common.Address, error) {
	return _RfqSettlement.Contract.Ftso(&_RfqSettlement.CallOpts)
}

// FtsoFeedId is a free data retrieval call binding the contract method 0x6caf711d.
//
// Solidity: function ftsoFeedId() view returns(bytes21)
func (_RfqSettlement *RfqSettlementCaller) FtsoFeedId(opts *bind.CallOpts) ([21]byte, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "ftsoFeedId")

	if err != nil {
		return *new([21]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([21]byte)).(*[21]byte)

	return out0, err

}

// FtsoFeedId is a free data retrieval call binding the contract method 0x6caf711d.
//
// Solidity: function ftsoFeedId() view returns(bytes21)
func (_RfqSettlement *RfqSettlementSession) FtsoFeedId() ([21]byte, error) {
	return _RfqSettlement.Contract.FtsoFeedId(&_RfqSettlement.CallOpts)
}

// FtsoFeedId is a free data retrieval call binding the contract method 0x6caf711d.
//
// Solidity: function ftsoFeedId() view returns(bytes21)
func (_RfqSettlement *RfqSettlementCallerSession) FtsoFeedId() ([21]byte, error) {
	return _RfqSettlement.Contract.FtsoFeedId(&_RfqSettlement.CallOpts)
}

// FtsoMaxStaleness is a free data retrieval call binding the contract method 0x8bc16d72.
//
// Solidity: function ftsoMaxStaleness() view returns(uint256)
func (_RfqSettlement *RfqSettlementCaller) FtsoMaxStaleness(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "ftsoMaxStaleness")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FtsoMaxStaleness is a free data retrieval call binding the contract method 0x8bc16d72.
//
// Solidity: function ftsoMaxStaleness() view returns(uint256)
func (_RfqSettlement *RfqSettlementSession) FtsoMaxStaleness() (*big.Int, error) {
	return _RfqSettlement.Contract.FtsoMaxStaleness(&_RfqSettlement.CallOpts)
}

// FtsoMaxStaleness is a free data retrieval call binding the contract method 0x8bc16d72.
//
// Solidity: function ftsoMaxStaleness() view returns(uint256)
func (_RfqSettlement *RfqSettlementCallerSession) FtsoMaxStaleness() (*big.Int, error) {
	return _RfqSettlement.Contract.FtsoMaxStaleness(&_RfqSettlement.CallOpts)
}

// FtsoToleranceBps is a free data retrieval call binding the contract method 0xbc39ed27.
//
// Solidity: function ftsoToleranceBps() view returns(uint256)
func (_RfqSettlement *RfqSettlementCaller) FtsoToleranceBps(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "ftsoToleranceBps")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FtsoToleranceBps is a free data retrieval call binding the contract method 0xbc39ed27.
//
// Solidity: function ftsoToleranceBps() view returns(uint256)
func (_RfqSettlement *RfqSettlementSession) FtsoToleranceBps() (*big.Int, error) {
	return _RfqSettlement.Contract.FtsoToleranceBps(&_RfqSettlement.CallOpts)
}

// FtsoToleranceBps is a free data retrieval call binding the contract method 0xbc39ed27.
//
// Solidity: function ftsoToleranceBps() view returns(uint256)
func (_RfqSettlement *RfqSettlementCallerSession) FtsoToleranceBps() (*big.Int, error) {
	return _RfqSettlement.Contract.FtsoToleranceBps(&_RfqSettlement.CallOpts)
}

// HashFill is a free data retrieval call binding the contract method 0x092622f0.
//
// Solidity: function hashFill((bytes32,address,address,uint8,uint256,uint256,uint256) fill) view returns(bytes32)
func (_RfqSettlement *RfqSettlementCaller) HashFill(opts *bind.CallOpts, fill RfqSettlementFill) ([32]byte, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "hashFill", fill)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// HashFill is a free data retrieval call binding the contract method 0x092622f0.
//
// Solidity: function hashFill((bytes32,address,address,uint8,uint256,uint256,uint256) fill) view returns(bytes32)
func (_RfqSettlement *RfqSettlementSession) HashFill(fill RfqSettlementFill) ([32]byte, error) {
	return _RfqSettlement.Contract.HashFill(&_RfqSettlement.CallOpts, fill)
}

// HashFill is a free data retrieval call binding the contract method 0x092622f0.
//
// Solidity: function hashFill((bytes32,address,address,uint8,uint256,uint256,uint256) fill) view returns(bytes32)
func (_RfqSettlement *RfqSettlementCallerSession) HashFill(fill RfqSettlementFill) ([32]byte, error) {
	return _RfqSettlement.Contract.HashFill(&_RfqSettlement.CallOpts, fill)
}

// IsAttestedSigner is a free data retrieval call binding the contract method 0x09eaa58f.
//
// Solidity: function isAttestedSigner(address ) view returns(bool)
func (_RfqSettlement *RfqSettlementCaller) IsAttestedSigner(opts *bind.CallOpts, arg0 common.Address) (bool, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "isAttestedSigner", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsAttestedSigner is a free data retrieval call binding the contract method 0x09eaa58f.
//
// Solidity: function isAttestedSigner(address ) view returns(bool)
func (_RfqSettlement *RfqSettlementSession) IsAttestedSigner(arg0 common.Address) (bool, error) {
	return _RfqSettlement.Contract.IsAttestedSigner(&_RfqSettlement.CallOpts, arg0)
}

// IsAttestedSigner is a free data retrieval call binding the contract method 0x09eaa58f.
//
// Solidity: function isAttestedSigner(address ) view returns(bool)
func (_RfqSettlement *RfqSettlementCallerSession) IsAttestedSigner(arg0 common.Address) (bool, error) {
	return _RfqSettlement.Contract.IsAttestedSigner(&_RfqSettlement.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RfqSettlement *RfqSettlementCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RfqSettlement *RfqSettlementSession) Owner() (common.Address, error) {
	return _RfqSettlement.Contract.Owner(&_RfqSettlement.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_RfqSettlement *RfqSettlementCallerSession) Owner() (common.Address, error) {
	return _RfqSettlement.Contract.Owner(&_RfqSettlement.CallOpts)
}

// QuoteToken is a free data retrieval call binding the contract method 0x217a4b70.
//
// Solidity: function quoteToken() view returns(address)
func (_RfqSettlement *RfqSettlementCaller) QuoteToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "quoteToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// QuoteToken is a free data retrieval call binding the contract method 0x217a4b70.
//
// Solidity: function quoteToken() view returns(address)
func (_RfqSettlement *RfqSettlementSession) QuoteToken() (common.Address, error) {
	return _RfqSettlement.Contract.QuoteToken(&_RfqSettlement.CallOpts)
}

// QuoteToken is a free data retrieval call binding the contract method 0x217a4b70.
//
// Solidity: function quoteToken() view returns(address)
func (_RfqSettlement *RfqSettlementCallerSession) QuoteToken() (common.Address, error) {
	return _RfqSettlement.Contract.QuoteToken(&_RfqSettlement.CallOpts)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_RfqSettlement *RfqSettlementCaller) Settled(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _RfqSettlement.contract.Call(opts, &out, "settled", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_RfqSettlement *RfqSettlementSession) Settled(arg0 [32]byte) (bool, error) {
	return _RfqSettlement.Contract.Settled(&_RfqSettlement.CallOpts, arg0)
}

// Settled is a free data retrieval call binding the contract method 0xd945af1d.
//
// Solidity: function settled(bytes32 ) view returns(bool)
func (_RfqSettlement *RfqSettlementCallerSession) Settled(arg0 [32]byte) (bool, error) {
	return _RfqSettlement.Contract.Settled(&_RfqSettlement.CallOpts, arg0)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RfqSettlement *RfqSettlementTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _RfqSettlement.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RfqSettlement *RfqSettlementSession) RenounceOwnership() (*types.Transaction, error) {
	return _RfqSettlement.Contract.RenounceOwnership(&_RfqSettlement.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_RfqSettlement *RfqSettlementTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _RfqSettlement.Contract.RenounceOwnership(&_RfqSettlement.TransactOpts)
}

// SetAttestedSigner is a paid mutator transaction binding the contract method 0x000b61c4.
//
// Solidity: function setAttestedSigner(address signer, bool allowed) returns()
func (_RfqSettlement *RfqSettlementTransactor) SetAttestedSigner(opts *bind.TransactOpts, signer common.Address, allowed bool) (*types.Transaction, error) {
	return _RfqSettlement.contract.Transact(opts, "setAttestedSigner", signer, allowed)
}

// SetAttestedSigner is a paid mutator transaction binding the contract method 0x000b61c4.
//
// Solidity: function setAttestedSigner(address signer, bool allowed) returns()
func (_RfqSettlement *RfqSettlementSession) SetAttestedSigner(signer common.Address, allowed bool) (*types.Transaction, error) {
	return _RfqSettlement.Contract.SetAttestedSigner(&_RfqSettlement.TransactOpts, signer, allowed)
}

// SetAttestedSigner is a paid mutator transaction binding the contract method 0x000b61c4.
//
// Solidity: function setAttestedSigner(address signer, bool allowed) returns()
func (_RfqSettlement *RfqSettlementTransactorSession) SetAttestedSigner(signer common.Address, allowed bool) (*types.Transaction, error) {
	return _RfqSettlement.Contract.SetAttestedSigner(&_RfqSettlement.TransactOpts, signer, allowed)
}

// SetFtsoBound is a paid mutator transaction binding the contract method 0xdc7c5899.
//
// Solidity: function setFtsoBound(address _ftso, bytes21 _feedId, uint256 _toleranceBps, uint256 _maxStaleness) returns()
func (_RfqSettlement *RfqSettlementTransactor) SetFtsoBound(opts *bind.TransactOpts, _ftso common.Address, _feedId [21]byte, _toleranceBps *big.Int, _maxStaleness *big.Int) (*types.Transaction, error) {
	return _RfqSettlement.contract.Transact(opts, "setFtsoBound", _ftso, _feedId, _toleranceBps, _maxStaleness)
}

// SetFtsoBound is a paid mutator transaction binding the contract method 0xdc7c5899.
//
// Solidity: function setFtsoBound(address _ftso, bytes21 _feedId, uint256 _toleranceBps, uint256 _maxStaleness) returns()
func (_RfqSettlement *RfqSettlementSession) SetFtsoBound(_ftso common.Address, _feedId [21]byte, _toleranceBps *big.Int, _maxStaleness *big.Int) (*types.Transaction, error) {
	return _RfqSettlement.Contract.SetFtsoBound(&_RfqSettlement.TransactOpts, _ftso, _feedId, _toleranceBps, _maxStaleness)
}

// SetFtsoBound is a paid mutator transaction binding the contract method 0xdc7c5899.
//
// Solidity: function setFtsoBound(address _ftso, bytes21 _feedId, uint256 _toleranceBps, uint256 _maxStaleness) returns()
func (_RfqSettlement *RfqSettlementTransactorSession) SetFtsoBound(_ftso common.Address, _feedId [21]byte, _toleranceBps *big.Int, _maxStaleness *big.Int) (*types.Transaction, error) {
	return _RfqSettlement.Contract.SetFtsoBound(&_RfqSettlement.TransactOpts, _ftso, _feedId, _toleranceBps, _maxStaleness)
}

// Settle is a paid mutator transaction binding the contract method 0xc59b4af9.
//
// Solidity: function settle((bytes32,address,address,uint8,uint256,uint256,uint256) fill, bytes attestationSig) returns()
func (_RfqSettlement *RfqSettlementTransactor) Settle(opts *bind.TransactOpts, fill RfqSettlementFill, attestationSig []byte) (*types.Transaction, error) {
	return _RfqSettlement.contract.Transact(opts, "settle", fill, attestationSig)
}

// Settle is a paid mutator transaction binding the contract method 0xc59b4af9.
//
// Solidity: function settle((bytes32,address,address,uint8,uint256,uint256,uint256) fill, bytes attestationSig) returns()
func (_RfqSettlement *RfqSettlementSession) Settle(fill RfqSettlementFill, attestationSig []byte) (*types.Transaction, error) {
	return _RfqSettlement.Contract.Settle(&_RfqSettlement.TransactOpts, fill, attestationSig)
}

// Settle is a paid mutator transaction binding the contract method 0xc59b4af9.
//
// Solidity: function settle((bytes32,address,address,uint8,uint256,uint256,uint256) fill, bytes attestationSig) returns()
func (_RfqSettlement *RfqSettlementTransactorSession) Settle(fill RfqSettlementFill, attestationSig []byte) (*types.Transaction, error) {
	return _RfqSettlement.Contract.Settle(&_RfqSettlement.TransactOpts, fill, attestationSig)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RfqSettlement *RfqSettlementTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _RfqSettlement.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RfqSettlement *RfqSettlementSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _RfqSettlement.Contract.TransferOwnership(&_RfqSettlement.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_RfqSettlement *RfqSettlementTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _RfqSettlement.Contract.TransferOwnership(&_RfqSettlement.TransactOpts, newOwner)
}

// RfqSettlementAttestedSignerSetIterator is returned from FilterAttestedSignerSet and is used to iterate over the raw logs and unpacked data for AttestedSignerSet events raised by the RfqSettlement contract.
type RfqSettlementAttestedSignerSetIterator struct {
	Event *RfqSettlementAttestedSignerSet // Event containing the contract specifics and raw log

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
func (it *RfqSettlementAttestedSignerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RfqSettlementAttestedSignerSet)
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
		it.Event = new(RfqSettlementAttestedSignerSet)
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
func (it *RfqSettlementAttestedSignerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RfqSettlementAttestedSignerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RfqSettlementAttestedSignerSet represents a AttestedSignerSet event raised by the RfqSettlement contract.
type RfqSettlementAttestedSignerSet struct {
	Signer  common.Address
	Allowed bool
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterAttestedSignerSet is a free log retrieval operation binding the contract event 0x931106548cad731249c18dd89c844dd95dd50ab7ac0d47157e609c8b15206598.
//
// Solidity: event AttestedSignerSet(address indexed signer, bool allowed)
func (_RfqSettlement *RfqSettlementFilterer) FilterAttestedSignerSet(opts *bind.FilterOpts, signer []common.Address) (*RfqSettlementAttestedSignerSetIterator, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _RfqSettlement.contract.FilterLogs(opts, "AttestedSignerSet", signerRule)
	if err != nil {
		return nil, err
	}
	return &RfqSettlementAttestedSignerSetIterator{contract: _RfqSettlement.contract, event: "AttestedSignerSet", logs: logs, sub: sub}, nil
}

// WatchAttestedSignerSet is a free log subscription operation binding the contract event 0x931106548cad731249c18dd89c844dd95dd50ab7ac0d47157e609c8b15206598.
//
// Solidity: event AttestedSignerSet(address indexed signer, bool allowed)
func (_RfqSettlement *RfqSettlementFilterer) WatchAttestedSignerSet(opts *bind.WatchOpts, sink chan<- *RfqSettlementAttestedSignerSet, signer []common.Address) (event.Subscription, error) {

	var signerRule []interface{}
	for _, signerItem := range signer {
		signerRule = append(signerRule, signerItem)
	}

	logs, sub, err := _RfqSettlement.contract.WatchLogs(opts, "AttestedSignerSet", signerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RfqSettlementAttestedSignerSet)
				if err := _RfqSettlement.contract.UnpackLog(event, "AttestedSignerSet", log); err != nil {
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

// ParseAttestedSignerSet is a log parse operation binding the contract event 0x931106548cad731249c18dd89c844dd95dd50ab7ac0d47157e609c8b15206598.
//
// Solidity: event AttestedSignerSet(address indexed signer, bool allowed)
func (_RfqSettlement *RfqSettlementFilterer) ParseAttestedSignerSet(log types.Log) (*RfqSettlementAttestedSignerSet, error) {
	event := new(RfqSettlementAttestedSignerSet)
	if err := _RfqSettlement.contract.UnpackLog(event, "AttestedSignerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RfqSettlementEIP712DomainChangedIterator is returned from FilterEIP712DomainChanged and is used to iterate over the raw logs and unpacked data for EIP712DomainChanged events raised by the RfqSettlement contract.
type RfqSettlementEIP712DomainChangedIterator struct {
	Event *RfqSettlementEIP712DomainChanged // Event containing the contract specifics and raw log

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
func (it *RfqSettlementEIP712DomainChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RfqSettlementEIP712DomainChanged)
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
		it.Event = new(RfqSettlementEIP712DomainChanged)
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
func (it *RfqSettlementEIP712DomainChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RfqSettlementEIP712DomainChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RfqSettlementEIP712DomainChanged represents a EIP712DomainChanged event raised by the RfqSettlement contract.
type RfqSettlementEIP712DomainChanged struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterEIP712DomainChanged is a free log retrieval operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_RfqSettlement *RfqSettlementFilterer) FilterEIP712DomainChanged(opts *bind.FilterOpts) (*RfqSettlementEIP712DomainChangedIterator, error) {

	logs, sub, err := _RfqSettlement.contract.FilterLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return &RfqSettlementEIP712DomainChangedIterator{contract: _RfqSettlement.contract, event: "EIP712DomainChanged", logs: logs, sub: sub}, nil
}

// WatchEIP712DomainChanged is a free log subscription operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_RfqSettlement *RfqSettlementFilterer) WatchEIP712DomainChanged(opts *bind.WatchOpts, sink chan<- *RfqSettlementEIP712DomainChanged) (event.Subscription, error) {

	logs, sub, err := _RfqSettlement.contract.WatchLogs(opts, "EIP712DomainChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RfqSettlementEIP712DomainChanged)
				if err := _RfqSettlement.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
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

// ParseEIP712DomainChanged is a log parse operation binding the contract event 0x0a6387c9ea3628b88a633bb4f3b151770f70085117a15f9bf3787cda53f13d31.
//
// Solidity: event EIP712DomainChanged()
func (_RfqSettlement *RfqSettlementFilterer) ParseEIP712DomainChanged(log types.Log) (*RfqSettlementEIP712DomainChanged, error) {
	event := new(RfqSettlementEIP712DomainChanged)
	if err := _RfqSettlement.contract.UnpackLog(event, "EIP712DomainChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RfqSettlementFilledIterator is returned from FilterFilled and is used to iterate over the raw logs and unpacked data for Filled events raised by the RfqSettlement contract.
type RfqSettlementFilledIterator struct {
	Event *RfqSettlementFilled // Event containing the contract specifics and raw log

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
func (it *RfqSettlementFilledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RfqSettlementFilled)
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
		it.Event = new(RfqSettlementFilled)
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
func (it *RfqSettlementFilledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RfqSettlementFilledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RfqSettlementFilled represents a Filled event raised by the RfqSettlement contract.
type RfqSettlementFilled struct {
	RfqId [32]byte
	Taker common.Address
	Maker common.Address
	Side  uint8
	Size  *big.Int
	Price *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterFilled is a free log retrieval operation binding the contract event 0xac04f0a075fe40c73978a88b555c53a8e47f9ade626bc2414af7413102a12377.
//
// Solidity: event Filled(bytes32 indexed rfqId, address indexed taker, address indexed maker, uint8 side, uint256 size, uint256 price)
func (_RfqSettlement *RfqSettlementFilterer) FilterFilled(opts *bind.FilterOpts, rfqId [][32]byte, taker []common.Address, maker []common.Address) (*RfqSettlementFilledIterator, error) {

	var rfqIdRule []interface{}
	for _, rfqIdItem := range rfqId {
		rfqIdRule = append(rfqIdRule, rfqIdItem)
	}
	var takerRule []interface{}
	for _, takerItem := range taker {
		takerRule = append(takerRule, takerItem)
	}
	var makerRule []interface{}
	for _, makerItem := range maker {
		makerRule = append(makerRule, makerItem)
	}

	logs, sub, err := _RfqSettlement.contract.FilterLogs(opts, "Filled", rfqIdRule, takerRule, makerRule)
	if err != nil {
		return nil, err
	}
	return &RfqSettlementFilledIterator{contract: _RfqSettlement.contract, event: "Filled", logs: logs, sub: sub}, nil
}

// WatchFilled is a free log subscription operation binding the contract event 0xac04f0a075fe40c73978a88b555c53a8e47f9ade626bc2414af7413102a12377.
//
// Solidity: event Filled(bytes32 indexed rfqId, address indexed taker, address indexed maker, uint8 side, uint256 size, uint256 price)
func (_RfqSettlement *RfqSettlementFilterer) WatchFilled(opts *bind.WatchOpts, sink chan<- *RfqSettlementFilled, rfqId [][32]byte, taker []common.Address, maker []common.Address) (event.Subscription, error) {

	var rfqIdRule []interface{}
	for _, rfqIdItem := range rfqId {
		rfqIdRule = append(rfqIdRule, rfqIdItem)
	}
	var takerRule []interface{}
	for _, takerItem := range taker {
		takerRule = append(takerRule, takerItem)
	}
	var makerRule []interface{}
	for _, makerItem := range maker {
		makerRule = append(makerRule, makerItem)
	}

	logs, sub, err := _RfqSettlement.contract.WatchLogs(opts, "Filled", rfqIdRule, takerRule, makerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RfqSettlementFilled)
				if err := _RfqSettlement.contract.UnpackLog(event, "Filled", log); err != nil {
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

// ParseFilled is a log parse operation binding the contract event 0xac04f0a075fe40c73978a88b555c53a8e47f9ade626bc2414af7413102a12377.
//
// Solidity: event Filled(bytes32 indexed rfqId, address indexed taker, address indexed maker, uint8 side, uint256 size, uint256 price)
func (_RfqSettlement *RfqSettlementFilterer) ParseFilled(log types.Log) (*RfqSettlementFilled, error) {
	event := new(RfqSettlementFilled)
	if err := _RfqSettlement.contract.UnpackLog(event, "Filled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RfqSettlementFtsoBoundSetIterator is returned from FilterFtsoBoundSet and is used to iterate over the raw logs and unpacked data for FtsoBoundSet events raised by the RfqSettlement contract.
type RfqSettlementFtsoBoundSetIterator struct {
	Event *RfqSettlementFtsoBoundSet // Event containing the contract specifics and raw log

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
func (it *RfqSettlementFtsoBoundSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RfqSettlementFtsoBoundSet)
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
		it.Event = new(RfqSettlementFtsoBoundSet)
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
func (it *RfqSettlementFtsoBoundSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RfqSettlementFtsoBoundSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RfqSettlementFtsoBoundSet represents a FtsoBoundSet event raised by the RfqSettlement contract.
type RfqSettlementFtsoBoundSet struct {
	Ftso         common.Address
	FeedId       [21]byte
	ToleranceBps *big.Int
	MaxStaleness *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterFtsoBoundSet is a free log retrieval operation binding the contract event 0x7cd76c3b4d1de0ad85485760969ea02131ccc735fc5be1fb9648802777005eca.
//
// Solidity: event FtsoBoundSet(address ftso, bytes21 feedId, uint256 toleranceBps, uint256 maxStaleness)
func (_RfqSettlement *RfqSettlementFilterer) FilterFtsoBoundSet(opts *bind.FilterOpts) (*RfqSettlementFtsoBoundSetIterator, error) {

	logs, sub, err := _RfqSettlement.contract.FilterLogs(opts, "FtsoBoundSet")
	if err != nil {
		return nil, err
	}
	return &RfqSettlementFtsoBoundSetIterator{contract: _RfqSettlement.contract, event: "FtsoBoundSet", logs: logs, sub: sub}, nil
}

// WatchFtsoBoundSet is a free log subscription operation binding the contract event 0x7cd76c3b4d1de0ad85485760969ea02131ccc735fc5be1fb9648802777005eca.
//
// Solidity: event FtsoBoundSet(address ftso, bytes21 feedId, uint256 toleranceBps, uint256 maxStaleness)
func (_RfqSettlement *RfqSettlementFilterer) WatchFtsoBoundSet(opts *bind.WatchOpts, sink chan<- *RfqSettlementFtsoBoundSet) (event.Subscription, error) {

	logs, sub, err := _RfqSettlement.contract.WatchLogs(opts, "FtsoBoundSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RfqSettlementFtsoBoundSet)
				if err := _RfqSettlement.contract.UnpackLog(event, "FtsoBoundSet", log); err != nil {
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

// ParseFtsoBoundSet is a log parse operation binding the contract event 0x7cd76c3b4d1de0ad85485760969ea02131ccc735fc5be1fb9648802777005eca.
//
// Solidity: event FtsoBoundSet(address ftso, bytes21 feedId, uint256 toleranceBps, uint256 maxStaleness)
func (_RfqSettlement *RfqSettlementFilterer) ParseFtsoBoundSet(log types.Log) (*RfqSettlementFtsoBoundSet, error) {
	event := new(RfqSettlementFtsoBoundSet)
	if err := _RfqSettlement.contract.UnpackLog(event, "FtsoBoundSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// RfqSettlementOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the RfqSettlement contract.
type RfqSettlementOwnershipTransferredIterator struct {
	Event *RfqSettlementOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *RfqSettlementOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(RfqSettlementOwnershipTransferred)
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
		it.Event = new(RfqSettlementOwnershipTransferred)
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
func (it *RfqSettlementOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *RfqSettlementOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// RfqSettlementOwnershipTransferred represents a OwnershipTransferred event raised by the RfqSettlement contract.
type RfqSettlementOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RfqSettlement *RfqSettlementFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*RfqSettlementOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _RfqSettlement.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &RfqSettlementOwnershipTransferredIterator{contract: _RfqSettlement.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RfqSettlement *RfqSettlementFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *RfqSettlementOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _RfqSettlement.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(RfqSettlementOwnershipTransferred)
				if err := _RfqSettlement.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_RfqSettlement *RfqSettlementFilterer) ParseOwnershipTransferred(log types.Log) (*RfqSettlementOwnershipTransferred, error) {
	event := new(RfqSettlementOwnershipTransferred)
	if err := _RfqSettlement.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
