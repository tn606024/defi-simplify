// Package defisimplify7702 exposes imported Defi Simplify contract deployment
// identities, ABIs, parity vectors, and dynamic delegated-account execution.
//
// The package reads only checked-in artifacts. It never resolves deployment
// data or ABIs from a remote source at runtime. The executor trusts the reviewed
// immutable implementation address selected from those artifacts and checks
// pending EIP-7702 delegation before every submission. Artifact and Base-fork
// tests verify the recorded runtime code hash. Delegation remains installed
// until explicitly replaced or cleared, including after a reverted execution.
package defisimplify7702
