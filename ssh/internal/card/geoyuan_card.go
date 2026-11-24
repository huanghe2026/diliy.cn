// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package card

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

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
	_ = time.Tick
	_ = context.Background
)

// GeoyuanCardMetaData contains all meta data concerning the GeoyuanCard contract.
var GeoyuanCardMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"h3Index\",\"type\":\"uint64\"},{\"internalType\":\"uint256\",\"name\":\"existingTokenId\",\"type\":\"uint256\"}],\"name\":\"CellAlreadyMinted\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"EmptyTitle\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"NotApproved\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"NotOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroAddress\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint64\",\"name\":\"h3Index\",\"type\":\"uint64\"}],\"name\":\"Minted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"cellMinted\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"cellOf\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"h3Index\",\"type\":\"uint64\"}],\"name\":\"isCellMinted\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"metadata\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"title\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"color\",\"type\":\"string\"},{\"internalType\":\"uint64\",\"name\":\"h3Index\",\"type\":\"uint64\"},{\"internalType\":\"int32\",\"name\":\"latE4\",\"type\":\"int32\"},{\"internalType\":\"int32\",\"name\":\"lonE4\",\"type\":\"int32\"},{\"internalType\":\"uint32\",\"name\":\"res\",\"type\":\"uint32\"},{\"internalType\":\"uint40\",\"name\":\"timestamp\",\"type\":\"uint40\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"h3Index\",\"type\":\"uint64\"},{\"internalType\":\"string\",\"name\":\"title\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"description\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"color\",\"type\":\"string\"},{\"internalType\":\"int32\",\"name\":\"latE4\",\"type\":\"int32\"},{\"internalType\":\"int32\",\"name\":\"lonE4\",\"type\":\"int32\"},{\"internalType\":\"uint32\",\"name\":\"res\",\"type\":\"uint32\"}],\"name\":\"mint\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"}],\"name\":\"myCards\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"who\",\"type\":\"address\"},{\"internalType\":\"uint64[]\",\"name\":\"neighborCells\",\"type\":\"uint64[]\"}],\"name\":\"neighborMatch\",\"outputs\":[{\"internalType\":\"uint256[]\",\"name\":\"\",\"type\":\"uint256[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"ownerOf\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokenURI\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"tokensOfOwner\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"tokenId\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x6080604052348015600e575f80fd5b506117b98061001c5f395ff3fe608060405234801561000f575f80fd5b50600436106100ef575f3560e01c806368c179d0116100935780639ee86fb5116100635780639ee86fb5146102ba578063c87b56dd146102cd578063e3684e39146102e0578063fffa344c14610307575f80fd5b806368c179d01461020f57806370a082311461023957806382a65bad1461025857806395d89b4114610298575f80fd5b806318160ddd116100ce57806318160ddd1461019157806323b872dd146101a7578063403d5b7e146101bc5780636352211e146101cf575f80fd5b8062b68201146100f357806304e8bf191461013a57806306fdde031461015a575b5f80fd5b61011e610101366004611004565b6001600160401b03165f9081526003602052604090205480151591565b6040805192151583526020830191909152015b60405180910390f35b61014d61014836600461103a565b610326565b6040516101319190611053565b6101846040518060400160405280600b81526020016a11d95bde5d585b90d85c9960aa1b81525081565b60405161013191906110c3565b6101995f5481565b604051908152602001610131565b6101ba6101b53660046110d5565b6103f3565b005b61014d6101ca36600461110f565b610525565b6101f76101dd36600461118e565b60016020525f90815260409020546001600160a01b031681565b6040516001600160a01b039091168152602001610131565b61019961021d3660046111a5565b600660209081525f928352604080842090915290825290205481565b61019961024736600461103a565b60076020525f908152604090205481565b61028061026636600461118e565b60026020525f90815260409020546001600160401b031681565b6040516001600160401b039091168152602001610131565b6101846040518060400160405280600381526020016247594360e81b81525081565b6101996102c8366004611222565b6106b8565b6101846102db36600461118e565b610ac2565b6102f36102ee36600461118e565b610b59565b60405161013198979695949392919061130a565b610199610315366004611004565b60036020525f908152604090205481565b6001600160a01b0381165f90815260076020526040812054606091816001600160401b038111156103595761035961138b565b604051908082528060200260200182016040528015610382578160200160208202803683370190505b50905060015b8281116103eb576001600160a01b0385165f908152600660209081526040808320848452909152902054826103be6001846113b3565b815181106103ce576103ce6113cc565b6020908102919091010152806103e3816113e0565b915050610388565b509392505050565b5f818152600160205260409020546001600160a01b038481169116146104435760405163b9c0876160e01b8152600481018290526001600160a01b03841660248201526044015b60405180910390fd5b6001600160a01b03821661046a5760405163d92e233d60e01b815260040160405180910390fd5b5f81815260016020818152604080842080546001600160a01b0319166001600160a01b038881169190911790915587168452600790915282208054919290916104b49084906113b3565b90915550506001600160a01b0382165f9081526007602052604081208054600192906104e19084906113f8565b909155505060405181906001600160a01b0380851691908616907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef905f90a4505050565b60605f826001600160401b038111156105405761054061138b565b604051908082528060200260200182016040528015610569578160200160208202803683370190505b5090505f805b8481101561061e575f60035f88888581811061058d5761058d6113cc565b90506020020160208101906105a29190611004565b6001600160401b0316815260208101919091526040015f2054905080158015906105e557505f818152600160205260409020546001600160a01b03898116911614155b15610615578084846105f6816113e0565b955081518110610608576106086113cc565b6020026020010181815250505b5060010161056f565b505f816001600160401b038111156106385761063861138b565b604051908082528060200260200182016040528015610661578160200160208202803683370190505b5090505f5b828110156106ad57838181518110610680576106806113cc565b602002602001015182828151811061069a5761069a6113cc565b6020908102919091010152600101610666565b509695505050505050565b6001600160401b038a165f9081526003602052604081205415610710576001600160401b038b165f8181526003602052604090819020549051630d22619760e31b81526004810192909252602482015260440161043a565b5f899003610731576040516352ddcdef60e11b815260040160405180910390fd5b3361074f5760405163d92e233d60e01b815260040160405180910390fd5b5f808154600101919050819055505f5490503360015f8381526020019081526020015f205f6101000a8154816001600160a01b0302191690836001600160a01b031602179055508a60025f8381526020019081526020015f205f6101000a8154816001600160401b0302191690836001600160401b031602179055508060035f8d6001600160401b03166001600160401b031681526020019081526020015f2081905550600160075f336001600160a01b03166001600160a01b031681526020019081526020015f205f82825461082691906113f8565b9091555050335f908152600660209081526040808320600783528184205484528252918290208390558151610120601f8d018390049092028101820190925261010082018b815282918d908d90819085018382808284375f92019190915250505090825250604080516020601f8c018190048102820181019092528a815291810191908b908b90819084018382808284375f92019190915250505090825250604080516020601f8a01819004810282018101909252888152918101919089908990819084018382808284375f9201829052509385525050506001600160401b038e16602080840191909152600388810b6040808601919091529088900b606085015263ffffffff8716608085015264ffffffffff421660a09094019390935284825260059052208151819061095b908261148f565b5060208201516001820190610970908261148f565b5060408201516002820190610985908261148f565b5060608201516003919091018054608084015160a085015160c086015160e0909601516001600160401b039095166bffffffffffffffffffffffff1990931692909217600160401b63ffffffff928316021767ffffffffffffffff60601b1916600160601b9282169290920263ffffffff60801b191691909117600160801b91909416029290921764ffffffffff60a01b1916600160a01b64ffffffffff90921691909102179055610a3681610d4e565b5f82815260046020526040902090610a4e908261148f565b50604051819033905f907fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef908290a46040516001600160401b038c1690339083907f2061d6101087f0f32bdf9de6aeb2b152e1ca941a282422b1d9e4b7e2353878fe905f90a49a9950505050505050505050565b60046020525f908152604090208054610ada9061140b565b80601f0160208091040260200160405190810160405280929190818152602001828054610b069061140b565b8015610b515780601f10610b2857610100808354040283529160200191610b51565b820191905f5260205f20905b815481529060010190602001808311610b3457829003601f168201915b505050505081565b60056020525f9081526040902080548190610b739061140b565b80601f0160208091040260200160405190810160405280929190818152602001828054610b9f9061140b565b8015610bea5780601f10610bc157610100808354040283529160200191610bea565b820191905f5260205f20905b815481529060010190602001808311610bcd57829003601f168201915b505050505090806001018054610bff9061140b565b80601f0160208091040260200160405190810160405280929190818152602001828054610c2b9061140b565b8015610c765780601f10610c4d57610100808354040283529160200191610c76565b820191905f5260205f20905b815481529060010190602001808311610c5957829003601f168201915b505050505090806002018054610c8b9061140b565b80601f0160208091040260200160405190810160405280929190818152602001828054610cb79061140b565b8015610d025780601f10610cd957610100808354040283529160200191610d02565b820191905f5260205f20905b815481529060010190602001808311610ce557829003601f168201915b5050505060039283015491926001600160401b03831692600160401b8104820b9250600160601b810490910b90600160801b810463ffffffff1690600160a01b900464ffffffffff1688565b5f818152600560205260409020606090610d6783610e22565b6003820154829060018201906002830190610d8a906001600160401b0316610e22565b600380870154610da391600160401b909104900b610e22565b600380880154610dbc91600160601b909104900b610e22565b6003880154610dd790600160801b900463ffffffff16610e22565b6003890154610df390600160a01b900464ffffffffff16610e22565b604051602001610e0b999897969594939291906115ce565b604051602081830303815290604052915050919050565b6060815f03610e485750506040805180820190915260018152600360fc1b602082015290565b5f82128015610e5d57610e5a83611716565b92505b60408051604e808252608082019092525f91829182916020820181803683370190505090505b5f861315610ee957610e96600a87611744565b610ea19060306113f8565b60f81b8184610eaf816113e0565b955081518110610ec157610ec16113cc565b60200101906001600160f81b03191690815f1a905350610ee2600a87611757565b9550610e83565b8315610f2857602d60f81b8184610eff816113e0565b955081518110610f1157610f116113cc565b60200101906001600160f81b03191690815f1a9053505b5f836001600160401b03811115610f4157610f4161138b565b6040519080825280601f01601f191660200182016040528015610f6b576020820181803683370190505b5090505f92505b83831015610fdf57816001610f8785876113b3565b610f9191906113b3565b81518110610fa157610fa16113cc565b602001015160f81c60f81b818481518110610fbe57610fbe6113cc565b60200101906001600160f81b03191690815f1a905350600190920191610f72565b9695505050505050565b80356001600160401b0381168114610fff575f80fd5b919050565b5f60208284031215611014575f80fd5b61101d82610fe9565b9392505050565b80356001600160a01b0381168114610fff575f80fd5b5f6020828403121561104a575f80fd5b61101d82611024565b602080825282518282018190525f918401906040840190835b8181101561108a57835183526020938401939092019160010161106c565b509095945050505050565b5f81518084528060208401602086015e5f602082860101526020601f19601f83011685010191505092915050565b602081525f61101d6020830184611095565b5f805f606084860312156110e7575f80fd5b6110f084611024565b92506110fe60208501611024565b929592945050506040919091013590565b5f805f60408486031215611121575f80fd5b61112a84611024565b925060208401356001600160401b03811115611144575f80fd5b8401601f81018613611154575f80fd5b80356001600160401b03811115611169575f80fd5b8660208260051b840101111561117d575f80fd5b939660209190910195509293505050565b5f6020828403121561119e575f80fd5b5035919050565b5f80604083850312156111b6575f80fd5b6111bf83611024565b946020939093013593505050565b5f8083601f8401126111dd575f80fd5b5081356001600160401b038111156111f3575f80fd5b60208301915083602082850101111561120a575f80fd5b9250929050565b8035600381900b8114610fff575f80fd5b5f805f805f805f805f8060e08b8d03121561123b575f80fd5b6112448b610fe9565b995060208b01356001600160401b0381111561125e575f80fd5b61126a8d828e016111cd565b909a5098505060408b01356001600160401b03811115611288575f80fd5b6112948d828e016111cd565b90985096505060608b01356001600160401b038111156112b2575f80fd5b6112be8d828e016111cd565b90965094506112d1905060808c01611211565b92506112df60a08c01611211565b915060c08b013563ffffffff811681146112f7575f80fd5b809150509295989b9194979a5092959850565b61010081525f61131e61010083018b611095565b8281036020840152611330818b611095565b90508281036040840152611344818a611095565b6001600160401b039890981660608401525050600394850b60808201529290930b60a083015263ffffffff1660c082015264ffffffffff90911660e0909101529392505050565b634e487b7160e01b5f52604160045260245ffd5b634e487b7160e01b5f52601160045260245ffd5b818103818111156113c6576113c661139f565b92915050565b634e487b7160e01b5f52603260045260245ffd5b5f600182016113f1576113f161139f565b5060010190565b808201808211156113c6576113c661139f565b600181811c9082168061141f57607f821691505b60208210810361143d57634e487b7160e01b5f52602260045260245ffd5b50919050565b601f82111561148a57805f5260205f20601f840160051c810160208510156114685750805b601f840160051c820191505b81811015611487575f8155600101611474565b50505b505050565b81516001600160401b038111156114a8576114a861138b565b6114bc816114b6845461140b565b84611443565b6020601f8211600181146114ee575f83156114d75750848201515b5f19600385901b1c1916600184901b178455611487565b5f84815260208120601f198516915b8281101561151d57878501518255602094850194600190920191016114fd565b508482101561153a57868401515f19600387901b60f8161c191681555b50505050600190811b01905550565b5f81518060208401855e5f93019283525090919050565b5f815461156c8161140b565b6001821680156115835760018114611598576115c5565b60ff19831686528115158202860193506115c5565b845f5260205f205f5b838110156115bd578154888201526001909101906020016115a1565b505081860193505b50505092915050565b717b226e616d65223a2247656f7975616e202360701b81525f6115f4601283018c611549565b6a1116113a34ba3632911d1160a91b8152611612600b82018c611560565b701116113232b9b1b934b83a34b7b7111d1160791b81529050611638601182018b611560565b6a11161131b7b637b9111d1160a91b81529050611658600b82018a611560565b6c1116113419a4b73232bc111d1160991b8152905061167a600d820189611549565b671116113630ba111d60c11b815290506116976008820188611549565b6616113637b7111d60c91b815290506116b36007820187611549565b6616113932b9911d60c91b81529050600781016116d08187611549565b6d1116113a34b6b2b9ba30b6b8111d60911b8152915050600e81016117056116f88287611549565b607d60f81b815260010190565b9d9c50505050505050505050505050565b5f600160ff1b820161172a5761172a61139f565b505f0390565b634e487b7160e01b5f52601260045260245ffd5b5f8261175257611752611730565b500790565b5f8261176557611765611730565b600160ff1b82145f198414161561177e5761177e61139f565b50059056fea264697066735822122009997cc30869d1ba98e2a10a5e4cd868e6df7e69426dc9b7752320802f7389a964736f6c634300081a0033",
}

// GeoyuanCardABI is the input ABI used to generate the binding from.
// Deprecated: Use GeoyuanCardMetaData.ABI instead.
var GeoyuanCardABI = GeoyuanCardMetaData.ABI

// GeoyuanCardBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GeoyuanCardMetaData.Bin instead.
var GeoyuanCardBin = GeoyuanCardMetaData.Bin

// DeployGeoyuanCard deploys a new Ethereum contract, binding an instance of GeoyuanCard to it.
func DeployGeoyuanCard(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GeoyuanCard, error) {
	parsed, err := GeoyuanCardMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GeoyuanCardBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GeoyuanCard{GeoyuanCardCaller: GeoyuanCardCaller{contract: contract}, GeoyuanCardTransactor: GeoyuanCardTransactor{contract: contract}, GeoyuanCardFilterer: GeoyuanCardFilterer{contract: contract}}, nil
}

// GeoyuanCard is an auto generated Go binding around an Ethereum contract.
type GeoyuanCard struct {
	GeoyuanCardCaller     // Read-only binding to the contract
	GeoyuanCardTransactor // Write-only binding to the contract
	GeoyuanCardFilterer   // Log filterer for contract events
}

// GeoyuanCardCaller is an auto generated read-only Go binding around an Ethereum contract.
type GeoyuanCardCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GeoyuanCardTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GeoyuanCardTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GeoyuanCardFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GeoyuanCardFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GeoyuanCardSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GeoyuanCardSession struct {
	Contract     *GeoyuanCard      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GeoyuanCardCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GeoyuanCardCallerSession struct {
	Contract *GeoyuanCardCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// GeoyuanCardTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GeoyuanCardTransactorSession struct {
	Contract     *GeoyuanCardTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// GeoyuanCardRaw is an auto generated low-level Go binding around an Ethereum contract.
type GeoyuanCardRaw struct {
	Contract *GeoyuanCard // Generic contract binding to access the raw methods on
}

// GeoyuanCardCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GeoyuanCardCallerRaw struct {
	Contract *GeoyuanCardCaller // Generic read-only contract binding to access the raw methods on
}

// GeoyuanCardTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GeoyuanCardTransactorRaw struct {
	Contract *GeoyuanCardTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGeoyuanCard creates a new instance of GeoyuanCard, bound to a specific deployed contract.
func NewGeoyuanCard(address common.Address, backend bind.ContractBackend) (*GeoyuanCard, error) {
	contract, err := bindGeoyuanCard(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GeoyuanCard{GeoyuanCardCaller: GeoyuanCardCaller{contract: contract}, GeoyuanCardTransactor: GeoyuanCardTransactor{contract: contract}, GeoyuanCardFilterer: GeoyuanCardFilterer{contract: contract}}, nil
}

// NewGeoyuanCardCaller creates a new read-only instance of GeoyuanCard, bound to a specific deployed contract.
func NewGeoyuanCardCaller(address common.Address, caller bind.ContractCaller) (*GeoyuanCardCaller, error) {
	contract, err := bindGeoyuanCard(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GeoyuanCardCaller{contract: contract}, nil
}

// NewGeoyuanCardTransactor creates a new write-only instance of GeoyuanCard, bound to a specific deployed contract.
func NewGeoyuanCardTransactor(address common.Address, transactor bind.ContractTransactor) (*GeoyuanCardTransactor, error) {
	contract, err := bindGeoyuanCard(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GeoyuanCardTransactor{contract: contract}, nil
}

// NewGeoyuanCardFilterer creates a new log filterer instance of GeoyuanCard, bound to a specific deployed contract.
func NewGeoyuanCardFilterer(address common.Address, filterer bind.ContractFilterer) (*GeoyuanCardFilterer, error) {
	contract, err := bindGeoyuanCard(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GeoyuanCardFilterer{contract: contract}, nil
}

// bindGeoyuanCard binds a generic wrapper to an already deployed contract.
func bindGeoyuanCard(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GeoyuanCardMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GeoyuanCard *GeoyuanCardRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GeoyuanCard.Contract.GeoyuanCardCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GeoyuanCard *GeoyuanCardRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.GeoyuanCardTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GeoyuanCard *GeoyuanCardRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.GeoyuanCardTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GeoyuanCard *GeoyuanCardCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GeoyuanCard.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GeoyuanCard *GeoyuanCardTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GeoyuanCard *GeoyuanCardTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCaller) BalanceOf(opts *bind.CallOpts, arg0 common.Address) (*big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "balanceOf", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _GeoyuanCard.Contract.BalanceOf(&_GeoyuanCard.CallOpts, arg0)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCallerSession) BalanceOf(arg0 common.Address) (*big.Int, error) {
	return _GeoyuanCard.Contract.BalanceOf(&_GeoyuanCard.CallOpts, arg0)
}

// CellMinted is a free data retrieval call binding the contract method 0xfffa344c.
//
// Solidity: function cellMinted(uint64 ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCaller) CellMinted(opts *bind.CallOpts, arg0 uint64) (*big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "cellMinted", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CellMinted is a free data retrieval call binding the contract method 0xfffa344c.
//
// Solidity: function cellMinted(uint64 ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardSession) CellMinted(arg0 uint64) (*big.Int, error) {
	return _GeoyuanCard.Contract.CellMinted(&_GeoyuanCard.CallOpts, arg0)
}

// CellMinted is a free data retrieval call binding the contract method 0xfffa344c.
//
// Solidity: function cellMinted(uint64 ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCallerSession) CellMinted(arg0 uint64) (*big.Int, error) {
	return _GeoyuanCard.Contract.CellMinted(&_GeoyuanCard.CallOpts, arg0)
}

// CellOf is a free data retrieval call binding the contract method 0x82a65bad.
//
// Solidity: function cellOf(uint256 ) view returns(uint64)
func (_GeoyuanCard *GeoyuanCardCaller) CellOf(opts *bind.CallOpts, arg0 *big.Int) (uint64, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "cellOf", arg0)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// CellOf is a free data retrieval call binding the contract method 0x82a65bad.
//
// Solidity: function cellOf(uint256 ) view returns(uint64)
func (_GeoyuanCard *GeoyuanCardSession) CellOf(arg0 *big.Int) (uint64, error) {
	return _GeoyuanCard.Contract.CellOf(&_GeoyuanCard.CallOpts, arg0)
}

// CellOf is a free data retrieval call binding the contract method 0x82a65bad.
//
// Solidity: function cellOf(uint256 ) view returns(uint64)
func (_GeoyuanCard *GeoyuanCardCallerSession) CellOf(arg0 *big.Int) (uint64, error) {
	return _GeoyuanCard.Contract.CellOf(&_GeoyuanCard.CallOpts, arg0)
}

// IsCellMinted is a free data retrieval call binding the contract method 0x00b68201.
//
// Solidity: function isCellMinted(uint64 h3Index) view returns(bool, uint256)
func (_GeoyuanCard *GeoyuanCardCaller) IsCellMinted(opts *bind.CallOpts, h3Index uint64) (bool, *big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "isCellMinted", h3Index)

	if err != nil {
		return *new(bool), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// IsCellMinted is a free data retrieval call binding the contract method 0x00b68201.
//
// Solidity: function isCellMinted(uint64 h3Index) view returns(bool, uint256)
func (_GeoyuanCard *GeoyuanCardSession) IsCellMinted(h3Index uint64) (bool, *big.Int, error) {
	return _GeoyuanCard.Contract.IsCellMinted(&_GeoyuanCard.CallOpts, h3Index)
}

// IsCellMinted is a free data retrieval call binding the contract method 0x00b68201.
//
// Solidity: function isCellMinted(uint64 h3Index) view returns(bool, uint256)
func (_GeoyuanCard *GeoyuanCardCallerSession) IsCellMinted(h3Index uint64) (bool, *big.Int, error) {
	return _GeoyuanCard.Contract.IsCellMinted(&_GeoyuanCard.CallOpts, h3Index)
}

// Metadata is a free data retrieval call binding the contract method 0xe3684e39.
//
// Solidity: function metadata(uint256 ) view returns(string title, string description, string color, uint64 h3Index, int32 latE4, int32 lonE4, uint32 res, uint40 timestamp)
func (_GeoyuanCard *GeoyuanCardCaller) Metadata(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Title       string
	Description string
	Color       string
	H3Index     uint64
	LatE4       int32
	LonE4       int32
	Res         uint32
	Timestamp   *big.Int
}, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "metadata", arg0)

	outstruct := new(struct {
		Title       string
		Description string
		Color       string
		H3Index     uint64
		LatE4       int32
		LonE4       int32
		Res         uint32
		Timestamp   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Title = *abi.ConvertType(out[0], new(string)).(*string)
	outstruct.Description = *abi.ConvertType(out[1], new(string)).(*string)
	outstruct.Color = *abi.ConvertType(out[2], new(string)).(*string)
	outstruct.H3Index = *abi.ConvertType(out[3], new(uint64)).(*uint64)
	outstruct.LatE4 = *abi.ConvertType(out[4], new(int32)).(*int32)
	outstruct.LonE4 = *abi.ConvertType(out[5], new(int32)).(*int32)
	outstruct.Res = *abi.ConvertType(out[6], new(uint32)).(*uint32)
	outstruct.Timestamp = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Metadata is a free data retrieval call binding the contract method 0xe3684e39.
//
// Solidity: function metadata(uint256 ) view returns(string title, string description, string color, uint64 h3Index, int32 latE4, int32 lonE4, uint32 res, uint40 timestamp)
func (_GeoyuanCard *GeoyuanCardSession) Metadata(arg0 *big.Int) (struct {
	Title       string
	Description string
	Color       string
	H3Index     uint64
	LatE4       int32
	LonE4       int32
	Res         uint32
	Timestamp   *big.Int
}, error) {
	return _GeoyuanCard.Contract.Metadata(&_GeoyuanCard.CallOpts, arg0)
}

// Metadata is a free data retrieval call binding the contract method 0xe3684e39.
//
// Solidity: function metadata(uint256 ) view returns(string title, string description, string color, uint64 h3Index, int32 latE4, int32 lonE4, uint32 res, uint40 timestamp)
func (_GeoyuanCard *GeoyuanCardCallerSession) Metadata(arg0 *big.Int) (struct {
	Title       string
	Description string
	Color       string
	H3Index     uint64
	LatE4       int32
	LonE4       int32
	Res         uint32
	Timestamp   *big.Int
}, error) {
	return _GeoyuanCard.Contract.Metadata(&_GeoyuanCard.CallOpts, arg0)
}

// MyCards is a free data retrieval call binding the contract method 0x04e8bf19.
//
// Solidity: function myCards(address who) view returns(uint256[])
func (_GeoyuanCard *GeoyuanCardCaller) MyCards(opts *bind.CallOpts, who common.Address) ([]*big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "myCards", who)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// MyCards is a free data retrieval call binding the contract method 0x04e8bf19.
//
// Solidity: function myCards(address who) view returns(uint256[])
func (_GeoyuanCard *GeoyuanCardSession) MyCards(who common.Address) ([]*big.Int, error) {
	return _GeoyuanCard.Contract.MyCards(&_GeoyuanCard.CallOpts, who)
}

// MyCards is a free data retrieval call binding the contract method 0x04e8bf19.
//
// Solidity: function myCards(address who) view returns(uint256[])
func (_GeoyuanCard *GeoyuanCardCallerSession) MyCards(who common.Address) ([]*big.Int, error) {
	return _GeoyuanCard.Contract.MyCards(&_GeoyuanCard.CallOpts, who)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_GeoyuanCard *GeoyuanCardCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_GeoyuanCard *GeoyuanCardSession) Name() (string, error) {
	return _GeoyuanCard.Contract.Name(&_GeoyuanCard.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_GeoyuanCard *GeoyuanCardCallerSession) Name() (string, error) {
	return _GeoyuanCard.Contract.Name(&_GeoyuanCard.CallOpts)
}

// NeighborMatch is a free data retrieval call binding the contract method 0x403d5b7e.
//
// Solidity: function neighborMatch(address who, uint64[] neighborCells) view returns(uint256[])
func (_GeoyuanCard *GeoyuanCardCaller) NeighborMatch(opts *bind.CallOpts, who common.Address, neighborCells []uint64) ([]*big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "neighborMatch", who, neighborCells)

	if err != nil {
		return *new([]*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]*big.Int)).(*[]*big.Int)

	return out0, err

}

// NeighborMatch is a free data retrieval call binding the contract method 0x403d5b7e.
//
// Solidity: function neighborMatch(address who, uint64[] neighborCells) view returns(uint256[])
func (_GeoyuanCard *GeoyuanCardSession) NeighborMatch(who common.Address, neighborCells []uint64) ([]*big.Int, error) {
	return _GeoyuanCard.Contract.NeighborMatch(&_GeoyuanCard.CallOpts, who, neighborCells)
}

// NeighborMatch is a free data retrieval call binding the contract method 0x403d5b7e.
//
// Solidity: function neighborMatch(address who, uint64[] neighborCells) view returns(uint256[])
func (_GeoyuanCard *GeoyuanCardCallerSession) NeighborMatch(who common.Address, neighborCells []uint64) ([]*big.Int, error) {
	return _GeoyuanCard.Contract.NeighborMatch(&_GeoyuanCard.CallOpts, who, neighborCells)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 ) view returns(address)
func (_GeoyuanCard *GeoyuanCardCaller) OwnerOf(opts *bind.CallOpts, arg0 *big.Int) (common.Address, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "ownerOf", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 ) view returns(address)
func (_GeoyuanCard *GeoyuanCardSession) OwnerOf(arg0 *big.Int) (common.Address, error) {
	return _GeoyuanCard.Contract.OwnerOf(&_GeoyuanCard.CallOpts, arg0)
}

// OwnerOf is a free data retrieval call binding the contract method 0x6352211e.
//
// Solidity: function ownerOf(uint256 ) view returns(address)
func (_GeoyuanCard *GeoyuanCardCallerSession) OwnerOf(arg0 *big.Int) (common.Address, error) {
	return _GeoyuanCard.Contract.OwnerOf(&_GeoyuanCard.CallOpts, arg0)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_GeoyuanCard *GeoyuanCardCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_GeoyuanCard *GeoyuanCardSession) Symbol() (string, error) {
	return _GeoyuanCard.Contract.Symbol(&_GeoyuanCard.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_GeoyuanCard *GeoyuanCardCallerSession) Symbol() (string, error) {
	return _GeoyuanCard.Contract.Symbol(&_GeoyuanCard.CallOpts)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 ) view returns(string)
func (_GeoyuanCard *GeoyuanCardCaller) TokenURI(opts *bind.CallOpts, arg0 *big.Int) (string, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "tokenURI", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 ) view returns(string)
func (_GeoyuanCard *GeoyuanCardSession) TokenURI(arg0 *big.Int) (string, error) {
	return _GeoyuanCard.Contract.TokenURI(&_GeoyuanCard.CallOpts, arg0)
}

// TokenURI is a free data retrieval call binding the contract method 0xc87b56dd.
//
// Solidity: function tokenURI(uint256 ) view returns(string)
func (_GeoyuanCard *GeoyuanCardCallerSession) TokenURI(arg0 *big.Int) (string, error) {
	return _GeoyuanCard.Contract.TokenURI(&_GeoyuanCard.CallOpts, arg0)
}

// TokensOfOwner is a free data retrieval call binding the contract method 0x68c179d0.
//
// Solidity: function tokensOfOwner(address , uint256 ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCaller) TokensOfOwner(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "tokensOfOwner", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokensOfOwner is a free data retrieval call binding the contract method 0x68c179d0.
//
// Solidity: function tokensOfOwner(address , uint256 ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardSession) TokensOfOwner(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _GeoyuanCard.Contract.TokensOfOwner(&_GeoyuanCard.CallOpts, arg0, arg1)
}

// TokensOfOwner is a free data retrieval call binding the contract method 0x68c179d0.
//
// Solidity: function tokensOfOwner(address , uint256 ) view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCallerSession) TokensOfOwner(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _GeoyuanCard.Contract.TokensOfOwner(&_GeoyuanCard.CallOpts, arg0, arg1)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GeoyuanCard.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_GeoyuanCard *GeoyuanCardSession) TotalSupply() (*big.Int, error) {
	return _GeoyuanCard.Contract.TotalSupply(&_GeoyuanCard.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_GeoyuanCard *GeoyuanCardCallerSession) TotalSupply() (*big.Int, error) {
	return _GeoyuanCard.Contract.TotalSupply(&_GeoyuanCard.CallOpts)
}

// Mint is a paid mutator transaction binding the contract method 0x9ee86fb5.
//
// Solidity: function mint(uint64 h3Index, string title, string description, string color, int32 latE4, int32 lonE4, uint32 res) returns(uint256 tokenId)
func (_GeoyuanCard *GeoyuanCardTransactor) Mint(opts *bind.TransactOpts, h3Index uint64, title string, description string, color string, latE4 int32, lonE4 int32, res uint32) (*types.Transaction, error) {
	return _GeoyuanCard.contract.Transact(opts, "mint", h3Index, title, description, color, latE4, lonE4, res)
}

// Mint is a paid mutator transaction binding the contract method 0x9ee86fb5.
//
// Solidity: function mint(uint64 h3Index, string title, string description, string color, int32 latE4, int32 lonE4, uint32 res) returns(uint256 tokenId)
func (_GeoyuanCard *GeoyuanCardSession) Mint(h3Index uint64, title string, description string, color string, latE4 int32, lonE4 int32, res uint32) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.Mint(&_GeoyuanCard.TransactOpts, h3Index, title, description, color, latE4, lonE4, res)
}

// Mint is a paid mutator transaction binding the contract method 0x9ee86fb5.
//
// Solidity: function mint(uint64 h3Index, string title, string description, string color, int32 latE4, int32 lonE4, uint32 res) returns(uint256 tokenId)
func (_GeoyuanCard *GeoyuanCardTransactorSession) Mint(h3Index uint64, title string, description string, color string, latE4 int32, lonE4 int32, res uint32) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.Mint(&_GeoyuanCard.TransactOpts, h3Index, title, description, color, latE4, lonE4, res)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_GeoyuanCard *GeoyuanCardTransactor) TransferFrom(opts *bind.TransactOpts, from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _GeoyuanCard.contract.Transact(opts, "transferFrom", from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_GeoyuanCard *GeoyuanCardSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.TransferFrom(&_GeoyuanCard.TransactOpts, from, to, tokenId)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address from, address to, uint256 tokenId) returns()
func (_GeoyuanCard *GeoyuanCardTransactorSession) TransferFrom(from common.Address, to common.Address, tokenId *big.Int) (*types.Transaction, error) {
	return _GeoyuanCard.Contract.TransferFrom(&_GeoyuanCard.TransactOpts, from, to, tokenId)
}

// GeoyuanCardMintedIterator is returned from FilterMinted and is used to iterate over the raw logs and unpacked data for Minted events raised by the GeoyuanCard contract.
type GeoyuanCardMintedIterator struct {
	Event *GeoyuanCardMinted // Event containing the contract specifics and raw log

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
func (it *GeoyuanCardMintedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GeoyuanCardMinted)
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
		it.Event = new(GeoyuanCardMinted)
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
func (it *GeoyuanCardMintedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GeoyuanCardMintedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GeoyuanCardMinted represents a Minted event raised by the GeoyuanCard contract.
type GeoyuanCardMinted struct {
	TokenId *big.Int
	Owner   common.Address
	H3Index uint64
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMinted is a free log retrieval operation binding the contract event 0x2061d6101087f0f32bdf9de6aeb2b152e1ca941a282422b1d9e4b7e2353878fe.
//
// Solidity: event Minted(uint256 indexed tokenId, address indexed owner, uint64 indexed h3Index)
func (_GeoyuanCard *GeoyuanCardFilterer) FilterMinted(opts *bind.FilterOpts, tokenId []*big.Int, owner []common.Address, h3Index []uint64) (*GeoyuanCardMintedIterator, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var h3IndexRule []interface{}
	for _, h3IndexItem := range h3Index {
		h3IndexRule = append(h3IndexRule, h3IndexItem)
	}

	logs, sub, err := _GeoyuanCard.contract.FilterLogs(opts, "Minted", tokenIdRule, ownerRule, h3IndexRule)
	if err != nil {
		return nil, err
	}
	return &GeoyuanCardMintedIterator{contract: _GeoyuanCard.contract, event: "Minted", logs: logs, sub: sub}, nil
}

// WatchMinted is a free log subscription operation binding the contract event 0x2061d6101087f0f32bdf9de6aeb2b152e1ca941a282422b1d9e4b7e2353878fe.
//
// Solidity: event Minted(uint256 indexed tokenId, address indexed owner, uint64 indexed h3Index)
func (_GeoyuanCard *GeoyuanCardFilterer) WatchMinted(opts *bind.WatchOpts, sink chan<- *GeoyuanCardMinted, tokenId []*big.Int, owner []common.Address, h3Index []uint64) (event.Subscription, error) {

	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}
	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var h3IndexRule []interface{}
	for _, h3IndexItem := range h3Index {
		h3IndexRule = append(h3IndexRule, h3IndexItem)
	}

	logs, sub, err := _GeoyuanCard.contract.WatchLogs(opts, "Minted", tokenIdRule, ownerRule, h3IndexRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GeoyuanCardMinted)
				if err := _GeoyuanCard.contract.UnpackLog(event, "Minted", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
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

// ParseMinted is a log parse operation binding the contract event 0x2061d6101087f0f32bdf9de6aeb2b152e1ca941a282422b1d9e4b7e2353878fe.
//
// Solidity: event Minted(uint256 indexed tokenId, address indexed owner, uint64 indexed h3Index)
func (_GeoyuanCard *GeoyuanCardFilterer) ParseMinted(log types.Log) (*GeoyuanCardMinted, error) {
	event := new(GeoyuanCardMinted)
	if err := _GeoyuanCard.contract.UnpackLog(event, "Minted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GeoyuanCardTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the GeoyuanCard contract.
type GeoyuanCardTransferIterator struct {
	Event *GeoyuanCardTransfer // Event containing the contract specifics and raw log

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
func (it *GeoyuanCardTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GeoyuanCardTransfer)
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
		it.Event = new(GeoyuanCardTransfer)
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
func (it *GeoyuanCardTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GeoyuanCardTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GeoyuanCardTransfer represents a Transfer event raised by the GeoyuanCard contract.
type GeoyuanCardTransfer struct {
	From    common.Address
	To      common.Address
	TokenId *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_GeoyuanCard *GeoyuanCardFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address, tokenId []*big.Int) (*GeoyuanCardTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _GeoyuanCard.contract.FilterLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return &GeoyuanCardTransferIterator{contract: _GeoyuanCard.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_GeoyuanCard *GeoyuanCardFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *GeoyuanCardTransfer, from []common.Address, to []common.Address, tokenId []*big.Int) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}
	var tokenIdRule []interface{}
	for _, tokenIdItem := range tokenId {
		tokenIdRule = append(tokenIdRule, tokenIdItem)
	}

	logs, sub, err := _GeoyuanCard.contract.WatchLogs(opts, "Transfer", fromRule, toRule, tokenIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GeoyuanCardTransfer)
				if err := _GeoyuanCard.contract.UnpackLog(event, "Transfer", log); err != nil {
					// If the signature doesn't match, skip this log.
					if errors.Is(err, bind.ErrEventSignatureMismatch) {
						continue
					}
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 indexed tokenId)
func (_GeoyuanCard *GeoyuanCardFilterer) ParseTransfer(log types.Log) (*GeoyuanCardTransfer, error) {
	event := new(GeoyuanCardTransfer)
	if err := _GeoyuanCard.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
