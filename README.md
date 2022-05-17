# balcheck

Utility to check Emeris data against actual blockchain nodes, in order to report
inaccuracies.

## Checks

Available checks are:

- balance
- staking balance
- unstaking balances

## Usage

```sh
go run ./cmd/balcheck -addr cosmos1qymla9gh8z2cmrylt008hkre0gry6h92sxgazg
```

This command will launch the checks for all the chains available in Emeris.
