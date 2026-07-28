package defisimplify7702

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tn606024/defi-simplify/config"
)

var (
	ErrInvalidDeployment   = errors.New("invalid Defi Simplify contracts deployment")
	ErrUnsupportedChain    = errors.New("unsupported Defi Simplify contracts chain")
	ErrUnknownArtifact     = errors.New("unknown Defi Simplify contracts artifact")
	ErrRuntimeCodeMismatch = errors.New("runtime code does not match configured deployment")
)

type RuntimeIdentity struct {
	Name            string
	Address         common.Address
	RuntimeCodeHash common.Hash
	Version         string
}

func (identity RuntimeIdentity) VerifyRuntimeCode(code []byte) error {
	actual := crypto.Keccak256Hash(code)
	if actual != identity.RuntimeCodeHash {
		return fmt.Errorf(
			"%w: %s at %s: expected %s, got %s",
			ErrRuntimeCodeMismatch,
			identity.Name,
			identity.Address.Hex(),
			identity.RuntimeCodeHash.Hex(),
			actual.Hex(),
		)
	}
	return nil
}

type ContractDeployment struct {
	RuntimeIdentity
	ABIPath string
}

type SecurityMetadata struct {
	Experimental      bool
	Notice            string
	Status            string
	SecurityGuarantee bool
	TotalLossRisk     bool
}

type Deployment struct {
	Chain       config.Chain
	ChainID     int
	NetworkName string
	EntryPoint  RuntimeIdentity
	Security    SecurityMetadata

	contracts map[Contract]ContractDeployment
}

var deploymentPaths = map[int]string{
	8453: "deployments/base-v1.json",
}

func DeploymentForChain(chain config.Chain) (Deployment, error) {
	chainID, err := chain.ChainID()
	if err != nil {
		return Deployment{}, fmt.Errorf("%w: %v", ErrUnsupportedChain, err)
	}
	deploymentPath, ok := deploymentPaths[chainID]
	if !ok {
		return Deployment{}, fmt.Errorf("%w: chain ID %d", ErrUnsupportedChain, chainID)
	}
	data, err := files.ReadFile(deploymentPath)
	if err != nil {
		return Deployment{}, fmt.Errorf("read deployment for chain %d: %w", chainID, err)
	}
	return parseDeployment(chain, chainID, data)
}

func (deployment Deployment) Contract(contract Contract) (ContractDeployment, error) {
	value, ok := deployment.contracts[contract]
	if !ok {
		return ContractDeployment{}, fmt.Errorf("%w: contract %q", ErrUnknownArtifact, contract)
	}
	return value, nil
}

func (deployment Deployment) Contracts() []Contract {
	contracts := make([]Contract, 0, len(deployment.contracts))
	for contract := range deployment.contracts {
		contracts = append(contracts, contract)
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i] < contracts[j] })
	return contracts
}

type deploymentManifest struct {
	SchemaVersion int                         `json:"schemaVersion"`
	Artifacts     map[string]artifactManifest `json:"artifacts"`
	EntryPoint    runtimeManifest             `json:"entryPoint"`
	Network       networkManifest             `json:"network"`
	Security      securityManifest            `json:"security"`
}

type artifactManifest struct {
	ABIPath         string `json:"abiPath"`
	Address         string `json:"address"`
	ContractName    string `json:"contractName"`
	RuntimeCodeHash string `json:"runtimeCodeHash"`
}

type runtimeManifest struct {
	Address         string `json:"address"`
	RuntimeCodeHash string `json:"runtimeCodeHash"`
	Version         string `json:"version"`
}

type networkManifest struct {
	ChainID int    `json:"chainId"`
	Name    string `json:"name"`
}

type securityManifest struct {
	Experimental      bool   `json:"experimental"`
	Notice            string `json:"notice"`
	SecurityGuarantee bool   `json:"securityGuarantee"`
	Status            string `json:"status"`
	TotalLossRisk     bool   `json:"totalLossRisk"`
}

func parseDeployment(chain config.Chain, expectedChainID int, data []byte) (Deployment, error) {
	var manifest deploymentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Deployment{}, fmt.Errorf("%w: decode JSON: %v", ErrInvalidDeployment, err)
	}
	if manifest.SchemaVersion != 1 {
		return Deployment{}, fmt.Errorf(
			"%w: unsupported schema version %d",
			ErrInvalidDeployment,
			manifest.SchemaVersion,
		)
	}
	if manifest.Network.ChainID != expectedChainID {
		return Deployment{}, fmt.Errorf(
			"%w: expected chain ID %d, got %d",
			ErrInvalidDeployment,
			expectedChainID,
			manifest.Network.ChainID,
		)
	}
	if strings.TrimSpace(manifest.Network.Name) == "" {
		return Deployment{}, fmt.Errorf("%w: network name is empty", ErrInvalidDeployment)
	}

	entryPoint, err := runtimeIdentity(
		"EntryPoint",
		manifest.EntryPoint.Address,
		manifest.EntryPoint.RuntimeCodeHash,
		manifest.EntryPoint.Version,
	)
	if err != nil {
		return Deployment{}, err
	}
	if strings.TrimSpace(entryPoint.Version) == "" {
		return Deployment{}, fmt.Errorf("%w: EntryPoint version is empty", ErrInvalidDeployment)
	}

	contracts := make(map[Contract]ContractDeployment, len(contractABIPaths))
	for contract, expectedABIPath := range contractABIPaths {
		artifact, ok := manifest.Artifacts[string(contract)]
		if !ok {
			return Deployment{}, fmt.Errorf("%w: missing contract %s", ErrInvalidDeployment, contract)
		}
		if artifact.ContractName != string(contract) {
			return Deployment{}, fmt.Errorf(
				"%w: contract %s declares name %q",
				ErrInvalidDeployment,
				contract,
				artifact.ContractName,
			)
		}
		if artifact.ABIPath != expectedABIPath {
			return Deployment{}, fmt.Errorf(
				"%w: contract %s ABI path is %q, expected %q",
				ErrInvalidDeployment,
				contract,
				artifact.ABIPath,
				expectedABIPath,
			)
		}
		identity, err := runtimeIdentity(
			string(contract),
			artifact.Address,
			artifact.RuntimeCodeHash,
			"",
		)
		if err != nil {
			return Deployment{}, err
		}
		contracts[contract] = ContractDeployment{
			RuntimeIdentity: identity,
			ABIPath:         artifact.ABIPath,
		}
	}

	if strings.TrimSpace(manifest.Security.Status) == "" ||
		strings.TrimSpace(manifest.Security.Notice) == "" {
		return Deployment{}, fmt.Errorf("%w: security metadata is incomplete", ErrInvalidDeployment)
	}
	return Deployment{
		Chain:       chain,
		ChainID:     manifest.Network.ChainID,
		NetworkName: manifest.Network.Name,
		EntryPoint:  entryPoint,
		Security: SecurityMetadata{
			Experimental:      manifest.Security.Experimental,
			Notice:            manifest.Security.Notice,
			Status:            manifest.Security.Status,
			SecurityGuarantee: manifest.Security.SecurityGuarantee,
			TotalLossRisk:     manifest.Security.TotalLossRisk,
		},
		contracts: contracts,
	}, nil
}

func runtimeIdentity(name, addressValue, hashValue, version string) (RuntimeIdentity, error) {
	if !common.IsHexAddress(addressValue) {
		return RuntimeIdentity{}, fmt.Errorf(
			"%w: %s address %q is invalid",
			ErrInvalidDeployment,
			name,
			addressValue,
		)
	}
	address := common.HexToAddress(addressValue)
	if address == (common.Address{}) {
		return RuntimeIdentity{}, fmt.Errorf("%w: %s address is zero", ErrInvalidDeployment, name)
	}
	if len(hashValue) != 66 || !strings.HasPrefix(hashValue, "0x") ||
		len(common.FromHex(hashValue)) != common.HashLength {
		return RuntimeIdentity{}, fmt.Errorf(
			"%w: %s runtime code hash %q is invalid",
			ErrInvalidDeployment,
			name,
			hashValue,
		)
	}
	return RuntimeIdentity{
		Name:            name,
		Address:         address,
		RuntimeCodeHash: common.HexToHash(hashValue),
		Version:         version,
	}, nil
}
