package aaveaddressbook

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestParseExportRejectsUnreviewedShape(t *testing.T) {
	valid, err := os.ReadFile("testdata/aave-v3-base-export.json")
	if err != nil {
		t.Fatalf("read export fixture: %v", err)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "unknown field",
			data: []byte(strings.Replace(
				string(valid),
				`"chainId": 8453`,
				`"chainId": 8453, "unreviewed": true`,
				1,
			)),
		},
		{
			name: "trailing JSON",
			data: append(append([]byte(nil), valid...), []byte("{}")...),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseExport(test.data)
			if !errors.Is(err, ErrInvalidExport) {
				t.Fatalf("ParseExport() error = %v, want ErrInvalidExport", err)
			}
		})
	}
}

func TestParseExportForUsesCallerOwnedChainDefinition(t *testing.T) {
	base, err := os.ReadFile("testdata/aave-v3-base-export.json")
	if err != nil {
		t.Fatalf("read export fixture: %v", err)
	}
	ethereum := strings.Replace(string(base), `"export": "AaveV3Base"`, `"export": "AaveV3Ethereum"`, 1)
	ethereum = strings.Replace(ethereum, `"chainId": 8453`, `"chainId": 1`, 1)

	exported, err := ParseExportFor([]byte(ethereum), ExportDefinition{
		Name:    "AaveV3Ethereum",
		ChainID: 1,
	})
	if err != nil {
		t.Fatalf("ParseExportFor() error = %v", err)
	}
	if exported.Export != "AaveV3Ethereum" || exported.ChainID != 1 {
		t.Fatalf("parsed export = %s/%d, want AaveV3Ethereum/1", exported.Export, exported.ChainID)
	}
}
