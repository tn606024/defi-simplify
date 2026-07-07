package config

import "fmt"

type Chain int

const (
	Base Chain = iota
)

var ChainInfo = map[Chain]struct {
	Name    string
	ChainID int
}{
	Base: {"Base", 8453},
}

func ChainIDToChain(chainID int) (Chain, error) {
	for c, info := range ChainInfo {
		if info.ChainID == chainID {
			return c, nil
		}
	}
	return Base, fmt.Errorf("chain not found")
}

var GasTokenDecimals = map[Chain]uint8{
	Base: 18,
}

func (c Chain) Name() (string, error) {
	info, ok := ChainInfo[c]
	if !ok {
		return "", fmt.Errorf("unsupported chain name: %d", c)
	}
	return info.Name, nil
}

func (c Chain) ChainID() (int, error) {
	info, ok := ChainInfo[c]
	if !ok {
		return 0, fmt.Errorf("unsupported chain id: %d", c)
	}
	return info.ChainID, nil
}

func (c Chain) GasTokenDecimals() (uint8, error) {
	decimals, ok := GasTokenDecimals[c]
	if !ok {
		return 0, fmt.Errorf("unsupported gas token decimals for chain %d", c)
	}
	return decimals, nil
}
