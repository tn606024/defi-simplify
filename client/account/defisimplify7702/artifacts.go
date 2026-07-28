package defisimplify7702

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

const SourceRepository = "https://github.com/tn606024/defi-simplify-contracts"

type Contract string

const (
	AccountContract                     Contract = "DefiSimplify7702Account"
	FlowAssertionsContract              Contract = "FlowAssertions"
	StaticCallUint256AssertionsContract Contract = "StaticCallUint256Assertions"
)

type GoldenVector string

const (
	BaseAaveV3DynamicStrategyVector GoldenVector = "BaseAaveV3DynamicStrategy"
	BaseAaveV3FlashLifecycleVector  GoldenVector = "BaseAaveV3FlashLifecycle"
	CallbackExecutionVector         GoldenVector = "CallbackExecution"
	DynamicCalldataPatchingVector   GoldenVector = "DynamicCalldataPatching"
	DynamicExecutionVector          GoldenVector = "DynamicExecution"
)

var contractABIPaths = map[Contract]string{
	AccountContract:                     "abi/DefiSimplify7702Account.json",
	FlowAssertionsContract:              "abi/FlowAssertions.json",
	StaticCallUint256AssertionsContract: "abi/StaticCallUint256Assertions.json",
}

var goldenPaths = map[GoldenVector]string{
	BaseAaveV3DynamicStrategyVector: "abi/BaseAaveV3DynamicStrategy.golden.json",
	BaseAaveV3FlashLifecycleVector:  "abi/BaseAaveV3FlashLifecycle.golden.json",
	CallbackExecutionVector:         "abi/CallbackExecution.golden.json",
	DynamicCalldataPatchingVector:   "abi/DynamicCalldataPatching.golden.json",
	DynamicExecutionVector:          "abi/DynamicExecution.golden.json",
}

//go:embed abi/*.json deployments/*.json
var files embed.FS

func ABI(contract Contract) ([]byte, error) {
	path, ok := contractABIPaths[contract]
	if !ok {
		return nil, fmt.Errorf("%w: ABI %q", ErrUnknownArtifact, contract)
	}
	return readArtifact(path)
}

func Golden(vector GoldenVector) ([]byte, error) {
	path, ok := goldenPaths[vector]
	if !ok {
		return nil, fmt.Errorf("%w: golden vector %q", ErrUnknownArtifact, vector)
	}
	return readArtifact(path)
}

func VerifyArtifacts() error {
	for contract := range contractABIPaths {
		data, err := ABI(contract)
		if err != nil {
			return err
		}
		if _, err := abi.JSON(bytes.NewReader(data)); err != nil {
			return fmt.Errorf("parse %s ABI: %w", contract, err)
		}
	}
	for vector := range goldenPaths {
		data, err := Golden(vector)
		if err != nil {
			return err
		}
		if !json.Valid(data) {
			return fmt.Errorf("golden vector %s is not valid JSON", vector)
		}
	}
	return nil
}

func readArtifact(path string) ([]byte, error) {
	data, err := files.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read embedded artifact %s: %w", path, err)
	}
	return append([]byte(nil), data...), nil
}
