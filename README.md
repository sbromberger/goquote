# goquote
stock quotes from the command line

usage:
```
goquote -s [+|-]<sort field> ticker1 ticker2 ... tickerN
```

`<sort field>` is one of
* `symbol` / `sym` / `s`
* `latest` / `l`
* `open` / `op` / `o`
* `close` / `cl` / `c`
* `change` / `chg` / `ch`
* `changepct` / `chgpct` / `pctchg` / `chgp` / `pchg` / `chp`/ `pch` / `%`
* `time` / `t`
* `volume` / `vol` / `v`

(optional) use `+` or `-` to specify sort order.
