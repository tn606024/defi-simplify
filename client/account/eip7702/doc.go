// Package eip7702 manages EIP-7702 account delegation lifecycle.
//
// EIP-7702 delegation is persistent. A successful authorization writes the EOA
// code to a delegation indicator, 0xef0100 || implementationAddress, and that
// delegation remains until a later authorization changes it or clears it by
// authorizing the zero address. Reverting EVM execution does not automatically
// restore the previous delegation.
package eip7702
