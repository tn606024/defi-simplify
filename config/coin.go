package config

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tn606024/defi-simplify/helper"
)

type Coin int

const (
	USDC Coin = iota
	AUSDC
	AVDUSDC
	WETH
	AWETH
	AVDWETH
	CBETH
	USDB
	WSTETH
	WEETH
	CBBTC
	EZETH
	GHO
	WRSETH
	LBTC
	EURC
	AAVE
)

var CoinAddress = map[Chain]map[Coin]common.Address{
	Base: {
		USDC:    common.HexToAddress("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913"),
		AUSDC:   common.HexToAddress("0x4e65fE4DbA92790696d040ac24Aa414708F5c0AB"),
		AVDUSDC: common.HexToAddress("0x59dca05b6c26dbd64b5381374aAaC5CD05644C28"),
		WETH:    common.HexToAddress("0x4200000000000000000000000000000000000006"),
		AWETH:   common.HexToAddress("0xD4a0e0b9149BCee3C920d2E00b5dE09138fd8bb7"),
		AVDWETH: common.HexToAddress("0x24e6e0795b3c7c71D965fCc4f371803d1c1DcA1E"),
		CBETH:   common.HexToAddress("0x2Ae3F1Ec7F1F5012CFEab0185bfc7aa3cf0DEc22"),
		USDB:    common.HexToAddress("0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA"),
		WSTETH:  common.HexToAddress("0xc1CBa3fCea344f92D9239c08C0568f6F2F0ee452"),
		WEETH:   common.HexToAddress("0x04C0599Ae5A44757c0af6F9eC3b93da8976c150A"),
		CBBTC:   common.HexToAddress("0xcbB7C0000aB88B473b1f5aFd9ef808440eed33Bf"),
		EZETH:   common.HexToAddress("0x2416092f143378750bb29b79eD961ab195CcEea5"),
		GHO:     common.HexToAddress("0x6Bb7a212910682DCFdbd5BCBb3e28FB4E8da10Ee"),
		WRSETH:  common.HexToAddress("0xEDfa23602D0EC14714057867A78d01e94176BEA0"),
		LBTC:    common.HexToAddress("0xecAc9C5F704e954931349Da37F60E39f515c11c1"),
		EURC:    common.HexToAddress("0x60a3E35Cc302bFA44Cb288Bc5a4F316Fdb1adb42"),
		AAVE:    common.HexToAddress("0x63706e401c06ac8513145b7687A14804d17f814b"),
	},
}

func AddressToCoin(chain Chain, address common.Address) (Coin, error) {
	coinAddresses, ok := CoinAddress[chain]
	if !ok {
		return 0, fmt.Errorf("unsupported coin address config for chain %d", chain)
	}
	for coin, addr := range coinAddresses {
		if addr == address {
			return coin, nil
		}
	}
	return 0, fmt.Errorf("coin not found")
}

func (c Coin) Address(chain Chain) (common.Address, error) {
	coinAddresses, ok := CoinAddress[chain]
	if !ok {
		return common.Address{}, fmt.Errorf("unsupported coin address config for chain %d", chain)
	}
	address, ok := coinAddresses[c]
	if !ok || address == (common.Address{}) {
		return common.Address{}, fmt.Errorf("unsupported coin address for coin %d on chain %d", c, chain)
	}
	return address, nil
}

var CoinDecimals = map[Coin]uint8{
	USDC:    6,
	AUSDC:   6,
	AVDUSDC: 6,
	WETH:    18,
	AWETH:   18,
	AVDWETH: 18,
	CBETH:   18,
	USDB:    6,
	WSTETH:  18,
	WEETH:   18,
	CBBTC:   8,
	EZETH:   18,
	GHO:     18,
	WRSETH:  18,
	LBTC:    8,
	EURC:    6,
	AAVE:    18,
}

func (c Coin) Decimals() (uint8, error) {
	decimals, ok := CoinDecimals[c]
	if !ok {
		return 0, fmt.Errorf("unsupported coin decimals for coin %d", c)
	}
	return decimals, nil
}

var CoinName = map[Coin]map[Chain]string{
	USDC: {
		Base: "USD Coin",
	},
	AUSDC: {
		Base: "Aave Base USDC",
	},
	AVDUSDC: {
		Base: "Aave Base Variable Debt USDC",
	},
	WETH: {
		Base: "Wrapped Ether",
	},
	AVDWETH: {
		Base: "Aave Base Variable Debt WETH",
	},
	AWETH: {
		Base: "Aave Base WETH",
	},
	CBETH: {
		Base: "Coinbase Wrapped Staked ETH",
	},
	USDB: {
		Base: "USD Base Coin",
	},
	WSTETH: {
		Base: "Wrapped liquid staked Ether 2.0",
	},
	WEETH: {
		Base: "Wrapped eETH",
	},
	CBBTC: {
		Base: "Coinbase Wrapped BTC",
	},
	EZETH: {
		Base: "Renzo Restaked ETH",
	},
	GHO: {
		Base: "Gho Token",
	},
	WRSETH: {
		Base: "rsETHWrapper",
	},
	LBTC: {
		Base: "Lombard Staked BTC",
	},
	EURC: {
		Base: "EURC",
	},
	AAVE: {
		Base: "Aave Token",
	},
}

func (c Coin) Name(chain Chain) (string, error) {
	chainNames, ok := CoinName[c]
	if !ok {
		return "", fmt.Errorf("unsupported coin name for coin %d", c)
	}
	name, ok := chainNames[chain]
	if !ok || name == "" {
		return "", fmt.Errorf("unsupported coin name for coin %d on chain %d", c, chain)
	}
	return name, nil
}

var CoinPermitSupported = map[Coin]map[Chain]bool{
	USDC:  {Base: true},
	AUSDC: {Base: true},
	WETH:  {Base: false},
	AWETH: {Base: true},
}

func (c Coin) PermitSupported(chain Chain) (bool, error) {
	chainSupport, ok := CoinPermitSupported[c]
	if !ok {
		return false, fmt.Errorf("unsupported permit support config for coin %d", c)
	}
	supported, ok := chainSupport[chain]
	if !ok {
		return false, fmt.Errorf("unsupported permit support config for coin %d on chain %d", c, chain)
	}
	return supported, nil
}

var CoinPermitVersion = map[Coin]map[Chain]string{
	USDC:    {Base: "2"},
	AVDUSDC: {Base: "1"},
	AUSDC:   {Base: "1"},
	AWETH:   {Base: "1"},
	AVDWETH: {Base: "1"},
}

func (c Coin) PermitVersion(chain Chain) (string, error) {
	chainVersions, ok := CoinPermitVersion[c]
	if !ok {
		return "", fmt.Errorf("unsupported permit version for coin %d", c)
	}
	version, ok := chainVersions[chain]
	if !ok || version == "" {
		return "", fmt.Errorf("unsupported permit version for coin %d on chain %d", c, chain)
	}
	return version, nil
}

func (c Coin) PermitDomain(chain Chain) (*helper.EIP712Domain, error) {
	name, err := c.Name(chain)
	if err != nil {
		return nil, err
	}
	version, err := c.PermitVersion(chain)
	if err != nil {
		return nil, err
	}
	chainID, err := chain.ChainID()
	if err != nil {
		return nil, err
	}
	address, err := c.Address(chain)
	if err != nil {
		return nil, err
	}
	return helper.NewEIP712Domain(name, version, big.NewInt(int64(chainID)), address), nil
}

var CoinAToken = map[Coin]Coin{
	USDC:  AUSDC,
	AUSDC: AUSDC,
	WETH:  AWETH,
	AWETH: AWETH,
}

func (c Coin) AToken() (Coin, error) {
	aToken, ok := CoinAToken[c]
	if !ok {
		return 0, fmt.Errorf("unsupported aToken for coin %d", c)
	}
	return aToken, nil
}

var CoinDebtToken = map[Coin]Coin{
	USDC:    AVDUSDC,
	AUSDC:   AVDUSDC,
	AVDUSDC: AVDUSDC,
	WETH:    AVDWETH,
	AWETH:   AVDWETH,
	AVDWETH: AVDWETH,
}

func (c Coin) DebtToken() (Coin, error) {
	debtToken, ok := CoinDebtToken[c]
	if !ok {
		return 0, fmt.Errorf("unsupported debt token for coin %d", c)
	}
	return debtToken, nil
}

var isDebtToken = map[Coin]bool{
	AVDUSDC: true,
	AVDWETH: true,
}

func (c Coin) IsDebtToken() bool {
	return isDebtToken[c]
}
