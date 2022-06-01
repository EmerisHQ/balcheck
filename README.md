# balcheck

Utility to check Emeris data against actual blockchain nodes, in order to report
inaccuracies.

Includes two binaries `balcheck` (runs http server)  and `balcheck-cli` (contains client command)

## Checks

Available checks are:

- balance
- staking balance
- unstaking balances

## `balcheck` usage

```sh
make balcheck
./build/balcheck
```

This command will start `balcheck` HTTP server (run default on `:8081`).

You can change listen address by passing `-listen-addr` argument:
```sh
./build/balcheck -listen-addr "localhost:8082"
```

Available endpoints are:
- /check/:address

Example: http://localhost:8081/check/cosmos1qymla9gh8z2cmrylt008hkre0gry6h92sxgazg

## `balcheck-cli` usage

```sh
make balcheck-cli
./build/balcheck-cli -addr cosmos1qymla9gh8z2cmrylt008hkre0gry6h92sxgazg
```

This command will launch the checks for all the chains available in Emeris.

