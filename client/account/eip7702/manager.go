package eip7702

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tn606024/defi-simplify/config"
)

type SetCodeClient interface {
	CodeReader
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
}

type Manager struct {
	client           SetCodeClient
	opts             *bind.TransactOpts
	authorizationKey *ecdsa.PrivateKey
	authority        common.Address
	chainID          *big.Int
}

func NewManager(client SetCodeClient, opts *bind.TransactOpts, authorizationKey *ecdsa.PrivateKey, chainID *big.Int) (*Manager, error) {
	if client == nil {
		return nil, errors.New("set-code client is nil")
	}
	if opts == nil {
		return nil, errors.New("transaction options are nil")
	}
	if opts.Signer == nil {
		return nil, errors.New("transaction signer is nil")
	}
	if authorizationKey == nil {
		return nil, errors.New("authorization key is nil")
	}
	if chainID == nil || chainID.Sign() < 0 {
		return nil, errors.New("chain ID must be non-negative")
	}

	return &Manager{
		client:           client,
		opts:             opts,
		authorizationKey: authorizationKey,
		authority:        crypto.PubkeyToAddress(authorizationKey.PublicKey),
		chainID:          new(big.Int).Set(chainID),
	}, nil
}

func (m *Manager) Delegate(ctx context.Context, implementation common.Address) (*types.Transaction, error) {
	return m.setDelegation(ctx, implementation)
}

func (m *Manager) DelegateToSimple7702(ctx context.Context, chain config.Chain) (*types.Transaction, error) {
	implementation, err := chain.Simple7702AccountImplementationAddress()
	if err != nil {
		return nil, err
	}
	return m.Delegate(ctx, implementation)
}

func (m *Manager) Clear(ctx context.Context) (*types.Transaction, error) {
	return m.setDelegation(ctx, common.Address{})
}

func (m *Manager) State(ctx context.Context, account common.Address) (DelegationState, error) {
	return ReadDelegationState(ctx, m.client, account)
}

func (m *Manager) AssertClean(ctx context.Context, account common.Address) error {
	return AssertClean(ctx, m.client, account)
}

func (m *Manager) AssertDelegatedTo(ctx context.Context, account common.Address, implementation common.Address) error {
	return AssertDelegatedTo(ctx, m.client, account, implementation)
}

func (m *Manager) setDelegation(ctx context.Context, implementation common.Address) (*types.Transaction, error) {
	tx, err := m.BuildDelegationTransaction(ctx, implementation)
	if err != nil {
		return nil, err
	}
	if m.opts.NoSend {
		return tx, nil
	}
	if err := m.client.SendTransaction(ctx, tx); err != nil {
		return nil, fmt.Errorf("send set-code transaction: %w", err)
	}
	return tx, nil
}

func (m *Manager) BuildDelegationTransaction(ctx context.Context, implementation common.Address) (*types.Transaction, error) {
	txNonce, err := m.transactionNonce(ctx)
	if err != nil {
		return nil, err
	}
	authNonce, err := m.authorizationNonce(ctx, txNonce)
	if err != nil {
		return nil, err
	}
	auth, err := SignAuthorization(m.authorizationKey, m.chainID, implementation, authNonce)
	if err != nil {
		return nil, err
	}

	to := m.opts.From
	value := m.opts.Value
	gasTipCap, gasFeeCap, err := m.fees(ctx)
	if err != nil {
		return nil, err
	}
	gasLimit, err := m.gasLimit(ctx, to, value, nil, gasTipCap, gasFeeCap, []types.SetCodeAuthorization{auth})
	if err != nil {
		return nil, err
	}

	return BuildSetCodeTransaction(SetCodeTransactionRequest{
		From:       m.opts.From,
		Signer:     m.opts.Signer,
		ChainID:    m.chainID,
		Nonce:      txNonce,
		To:         to,
		Value:      value,
		Gas:        gasLimit,
		GasFeeCap:  gasFeeCap,
		GasTipCap:  gasTipCap,
		AccessList: m.opts.AccessList,
		AuthList:   []types.SetCodeAuthorization{auth},
	})
}

func (m *Manager) transactionNonce(ctx context.Context) (uint64, error) {
	if m.opts.Nonce != nil {
		if m.opts.Nonce.Sign() < 0 {
			return 0, errors.New("transaction nonce is negative")
		}
		return m.opts.Nonce.Uint64(), nil
	}
	nonce, err := m.client.PendingNonceAt(ctx, m.opts.From)
	if err != nil {
		return 0, fmt.Errorf("read transaction nonce: %w", err)
	}
	return nonce, nil
}

func (m *Manager) authorizationNonce(ctx context.Context, txNonce uint64) (uint64, error) {
	if m.authority == m.opts.From {
		// For same-sender delegation, geth increments the transaction sender nonce
		// before applying EIP-7702 authorizations, so the authorization must sign
		// txNonce + 1.
		if txNonce == math.MaxUint64 {
			return 0, errors.New("authorization nonce overflow")
		}
		return txNonce + 1, nil
	}
	nonce, err := m.client.PendingNonceAt(ctx, m.authority)
	if err != nil {
		return 0, fmt.Errorf("read authorization nonce: %w", err)
	}
	return nonce, nil
}

func (m *Manager) fees(ctx context.Context) (*big.Int, *big.Int, error) {
	if m.opts.GasPrice != nil {
		return m.opts.GasPrice, m.opts.GasPrice, nil
	}
	if m.opts.GasFeeCap != nil && m.opts.GasTipCap != nil {
		return m.opts.GasTipCap, m.opts.GasFeeCap, nil
	}
	if m.opts.GasFeeCap != nil || m.opts.GasTipCap != nil {
		return nil, nil, errors.New("both gas fee cap and gas tip cap must be set when overriding EIP-1559 fees")
	}

	tipCap, err := m.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("suggest gas tip cap: %w", err)
	}
	header, err := m.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("read latest header: %w", err)
	}
	if header.BaseFee == nil {
		return tipCap, new(big.Int).Set(tipCap), nil
	}
	feeCap := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
	feeCap.Add(feeCap, tipCap)
	return tipCap, feeCap, nil
}

func (m *Manager) gasLimit(ctx context.Context, to common.Address, value *big.Int, data []byte, gasTipCap *big.Int, gasFeeCap *big.Int, authList []types.SetCodeAuthorization) (uint64, error) {
	if m.opts.GasLimit != 0 {
		return m.opts.GasLimit, nil
	}
	if value == nil {
		value = big.NewInt(0)
	}
	estimated, err := m.client.EstimateGas(ctx, ethereum.CallMsg{
		From:       m.opts.From,
		To:         &to,
		GasFeeCap:  gasFeeCap,
		GasTipCap:  gasTipCap,
		Value:      value,
		Data:       data,
		AccessList: m.opts.AccessList,
	})
	if err != nil {
		return 0, fmt.Errorf("estimate gas: %w", err)
	}

	authorizationOverhead := uint64(len(authList)) * params.CallNewAccountGas
	if math.MaxUint64-estimated < authorizationOverhead {
		return 0, errors.New("gas limit overflow")
	}
	return estimated + authorizationOverhead, nil
}
