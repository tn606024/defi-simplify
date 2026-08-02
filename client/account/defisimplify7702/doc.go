// Package defisimplify7702 exposes imported Defi Simplify contract deployment
// identities, ABIs, parity vectors, and delegated-account execution.
//
// Deployment identity and ABIs come only from checked-in artifacts, never a
// remote deployment source. Applications can verify the selected immutable
// account implementation against its on-chain runtime code during
// initialization. The executor then checks pending EIP-7702 delegation before
// every submission. Delegation remains installed until explicitly replaced or
// cleared, including after a reverted execution.
package defisimplify7702
