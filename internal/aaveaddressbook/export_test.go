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
