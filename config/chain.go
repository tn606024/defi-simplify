package config

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

func ChainIDToChain(chainID int) Chain {
	for c, info := range ChainInfo {
		if info.ChainID == chainID {
			return c
		}
	}
	return c
}

var GasTokenDecimals = map[Chain]uint8{
	Base: 18,
}

func (c Chain) GasTokenDecimals() uint8 {
	return GasTokenDecimals[c]
}
