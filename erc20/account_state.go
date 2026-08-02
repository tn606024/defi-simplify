package erc20

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	binderc20 "github.com/tn606024/defi-simplify/bind/erc20"
	"github.com/tn606024/defi-simplify/token"
)

var (
	// ErrAccountStateRead is returned when an ERC20 account-state request is
	// invalid or cannot be completed at one block.
	ErrAccountStateRead = errors.New("read ERC20 account state")
)

// StateBackend is the minimal read-only Ethereum backend required for a
// block-pinned ERC20 account-state read.
type StateBackend interface {
	bind.ContractCaller
	bind.BlockHashContractCaller
	HeaderByNumber(context.Context, *big.Int) (*types.Header, error)
}

// AccountState is an immutable view of one account's token balance and
// allowance for one spender at one block.
type AccountState struct {
	asset       token.Token
	account     common.Address
	spender     common.Address
	blockNumber uint64
	blockHash   common.Hash
	balance     *big.Int
	allowance   *big.Int
}

// LoadAccountState reads balance and allowance for one resolved token against
// the same latest block hash. Amounts remain in the token's base units.
func LoadAccountState(
	ctx context.Context,
	backend StateBackend,
	asset token.Token,
	account common.Address,
	spender common.Address,
) (*AccountState, error) {
	if backend == nil {
		return nil, fmt.Errorf("%w: backend is nil", ErrAccountStateRead)
	}
	if err := asset.Validate(); err != nil {
		return nil, fmt.Errorf("%w: asset: %w", ErrAccountStateRead, err)
	}
	if account == (common.Address{}) {
		return nil, fmt.Errorf("%w: account is zero", ErrAccountStateRead)
	}
	if spender == (common.Address{}) {
		return nil, fmt.Errorf("%w: spender is zero", ErrAccountStateRead)
	}

	header, err := backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: resolve block: %w", ErrAccountStateRead, err)
	}
	if header == nil || header.Number == nil || !header.Number.IsUint64() {
		return nil, fmt.Errorf("%w: header has no uint64 block number", ErrAccountStateRead)
	}
	blockHash := header.Hash()
	if blockHash == (common.Hash{}) {
		return nil, fmt.Errorf("%w: block hash is zero", ErrAccountStateRead)
	}

	tokenContract, err := binderc20.NewErc20Caller(asset.Address(), backend)
	if err != nil {
		return nil, fmt.Errorf("%w: bind token: %w", ErrAccountStateRead, err)
	}
	callOpts := &bind.CallOpts{
		Context:     ctx,
		BlockNumber: new(big.Int).Set(header.Number),
		BlockHash:   blockHash,
	}
	balance, err := tokenContract.BalanceOf(callOpts, account)
	if err != nil {
		return nil, fmt.Errorf("%w: read balance: %w", ErrAccountStateRead, err)
	}
	allowance, err := tokenContract.Allowance(callOpts, account, spender)
	if err != nil {
		return nil, fmt.Errorf("%w: read allowance: %w", ErrAccountStateRead, err)
	}
	if balance == nil || allowance == nil {
		return nil, fmt.Errorf("%w: token returned nil amount", ErrAccountStateRead)
	}

	return &AccountState{
		asset:       asset,
		account:     account,
		spender:     spender,
		blockNumber: header.Number.Uint64(),
		blockHash:   blockHash,
		balance:     cloneStateAmount(balance),
		allowance:   cloneStateAmount(allowance),
	}, nil
}

// Asset returns the resolved token observed by this state read.
func (s *AccountState) Asset() token.Token {
	return s.asset
}

// Account returns the token owner whose state was read.
func (s *AccountState) Account() common.Address {
	return s.account
}

// Spender returns the allowance spender whose state was read.
func (s *AccountState) Spender() common.Address {
	return s.spender
}

// BlockNumber returns the block number shared by every value in this state.
func (s *AccountState) BlockNumber() uint64 {
	return s.blockNumber
}

// BlockHash returns the block hash shared by every value in this state.
func (s *AccountState) BlockHash() common.Hash {
	return s.blockHash
}

// Balance returns a copy of the account's token balance in base units.
func (s *AccountState) Balance() *big.Int {
	return cloneStateAmount(s.balance)
}

// Allowance returns a copy of the account's allowance in base units.
func (s *AccountState) Allowance() *big.Int {
	return cloneStateAmount(s.allowance)
}

func cloneStateAmount(value *big.Int) *big.Int {
	if value == nil {
		return nil
	}
	return new(big.Int).Set(value)
}
