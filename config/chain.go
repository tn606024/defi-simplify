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

func (c Chain) Name() string {
	return ChainInfo[c].Name
}

func (c Chain) ChainID() int {
	return ChainInfo[c].ChainID
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

func (c Chain) GasTokenDecimals() uint8 {
	return GasTokenDecimals[c]
}
