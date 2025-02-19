# Alien Signals for Go

<p align="center">
  <img src="mascot.png" width="384"><br>
<p>

The lightest signal library for Dart, ported from [stackblitz/alien-signals](https://github.com/stackblitz/alien-signals).

> [!TIP]
> `alien_signals` is the fastest signal library currently.

## Benchmarks against original TypeScript implementation
TS
```
| benchmark                | avg                                     | min         | p75         | p99         | max         |
| ------------------------ | --------------------------------------- | ----------- | ----------- | ----------- | ----------- |
| propagate: 1 * 1         | `879.40 ns/iter`                        | `863.08 ns` | `881.93 ns` | `931.41 ns` | `942.93 ns` |
| propagate: 1 * 10        | `  4.13 µs/iter`                        | `  4.09 µs` | `  4.15 µs` | `  4.19 µs` | `  4.19 µs` |
| propagate: 1 * 100       | ` 36.05 µs/iter`                        | ` 35.21 µs` | ` 36.56 µs` | ` 36.85 µs` | ` 37.08 µs` |
| propagate: 1 * 1000      | `363.06 µs/iter`                        | `337.79 µs` | `360.45 µs` | `602.32 µs` | `828.24 µs` |
| propagate: 1 * 10000     | error: Maximum call stack size exceeded |
| propagate: 10 * 1        | `  7.59 µs/iter`                        | `  7.50 µs` | `  7.64 µs` | `  7.68 µs` | `  7.75 µs` |
| propagate: 10 * 10       | ` 40.17 µs/iter`                        | ` 39.45 µs` | ` 40.42 µs` | ` 40.67 µs` | ` 40.90 µs` |
| propagate: 10 * 100      | `377.55 µs/iter`                        | `345.79 µs` | `364.73 µs` | `721.62 µs` | `825.40 µs` |
| propagate: 10 * 1000     | `  3.77 ms/iter`                        | `  3.52 ms` | `  3.88 ms` | `  4.55 ms` | `  4.90 ms` |
| propagate: 10 * 10000    | error: Maximum call stack size exceeded |
| propagate: 100 * 1       | ` 72.93 µs/iter`                        | ` 70.33 µs` | ` 72.41 µs` | ` 84.55 µs` | `269.86 µs` |
| propagate: 100 * 10      | `396.95 µs/iter`                        | `383.79 µs` | `397.78 µs` | `470.70 µs` | `667.30 µs` |
| propagate: 100 * 100     | `  3.61 ms/iter`                        | `  3.54 ms` | `  3.62 ms` | `  3.81 ms` | `  3.83 ms` |
| propagate: 100 * 1000    | ` 61.82 ms/iter`                        | ` 60.57 ms` | ` 62.03 ms` | ` 62.65 ms` | ` 64.07 ms` |
| propagate: 100 * 10000   | error: Maximum call stack size exceeded |
| propagate: 1000 * 1      | `726.04 µs/iter`                        | `707.31 µs` | `729.20 µs` | `895.32 µs` | `927.52 µs` |
| propagate: 1000 * 10     | `  4.58 ms/iter`                        | `  4.00 ms` | `  4.76 ms` | `  5.83 ms` | `  6.07 ms` |
| propagate: 1000 * 100    | ` 58.94 ms/iter`                        | ` 58.18 ms` | ` 59.15 ms` | ` 59.90 ms` | ` 60.59 ms` |
| propagate: 1000 * 1000   | `640.89 ms/iter`                        | `636.51 ms` | `641.96 ms` | `644.19 ms` | `648.65 ms` |
| propagate: 1000 * 10000  | error: Maximum call stack size exceeded |
| propagate: 10000 * 1     | `  8.18 ms/iter`                        | `  7.79 ms` | `  8.31 ms` | `  8.49 ms` | `  8.53 ms` |
| propagate: 10000 * 10    | ` 58.76 ms/iter`                        | ` 57.19 ms` | ` 58.67 ms` | ` 60.52 ms` | ` 62.19 ms` |
| propagate: 10000 * 100   | `569.88 ms/iter`                        | `559.99 ms` | `570.88 ms` | `578.03 ms` | `581.46 ms` |
| propagate: 10000 * 1000  | `   7.13 s/iter`                        | `   6.72 s` | `   7.28 s` | `   7.35 s` | `   7.56 s` |
| propagate: 10000 * 10000 | error: Maximum call stack size exceeded |
```

Go
```
-+--------------------------+--------------+--------------+--------------+-------------+-------------+
-| BENCHMARK                |          AVG |          MIN |          P75 |         P99 |         MAX |
-+--------------------------+--------------+--------------+--------------+-------------+-------------+
-| propagate: 1 * 1         |        245ns |        120ns |        170ns |     1.182µs |     1.182µs |
-| propagate: 1 * 10        |        913ns |        862ns |        892ns |     1.182µs |     1.182µs |
-| propagate: 1 * 100       |      6.614µs |      5.751µs |      7.114µs |     7.876µs |     7.876µs |
-| propagate: 1 * 1000      |     54.281µs |     49.355µs |     55.137µs |    67.259µs |    67.259µs |
-| propagate: 1 * 10000     |    523.035µs |    452.471µs |    533.979µs |   897.419µs |   897.419µs |
-| propagate: 10 * 1        |        860ns |        811ns |        822ns |     1.133µs |     1.133µs |
-| propagate: 10 * 10       |      5.511µs |      5.099µs |       5.25µs |     8.636µs |     8.636µs |
-| propagate: 10 * 100      |      46.36µs |     46.109µs |      46.45µs |    47.241µs |    47.241µs |
-| propagate: 10 * 1000     |    516.253µs |    460.517µs |    516.505µs |   698.095µs |   698.095µs |
-| propagate: 10 * 10000    |   5.307504ms |   4.902171ms |   5.430799ms |  6.032809ms |  6.032809ms |
-| propagate: 100 * 1       |      8.105µs |      8.015µs |      8.046µs |     8.656µs |     8.656µs |
-| propagate: 100 * 10      |      53.76µs |     51.169µs |     57.311µs |    57.952µs |    57.952µs |
-| propagate: 100 * 100     |      528.4µs |    518.578µs |    531.203µs |   542.485µs |   542.485µs |
-| propagate: 100 * 1000    |   5.657486ms |   5.045708ms |    6.07453ms |  6.945918ms |  6.945918ms |
-| propagate: 100 * 10000   |  54.179415ms |  52.110433ms |  56.017539ms | 56.933642ms | 56.933642ms |
-| propagate: 1000 * 1      |     76.831µs |     75.094µs |     75.475µs |    87.819µs |    87.819µs |
-| propagate: 1000 * 10     |    612.015µs |    529.319µs |    590.787µs |  1.006358ms |  1.006358ms |
-| propagate: 1000 * 100    |   6.532832ms |   5.988413ms |   6.685005ms |  6.873769ms |  6.873769ms |
-| propagate: 1000 * 1000   |  55.287091ms |   53.40665ms |  55.554066ms | 57.087619ms | 57.087619ms |
-| propagate: 1000 * 10000  | 533.734675ms | 524.460397ms | 539.883531ms | 543.99004ms | 543.99004ms |
-| propagate: 10000 * 1     |    781.496µs |    767.147µs |    788.399µs |   803.497µs |   803.497µs |
-| propagate: 10000 * 10    |   9.808555ms |   9.239706ms |   9.857205ms | 11.427108ms | 11.427108ms |
-| propagate: 10000 * 100   |  64.933085ms |  64.196962ms |  65.353068ms | 65.612147ms | 65.612147ms |
-| propagate: 10000 * 1000  | 533.553871ms | 528.541617ms | 535.905941ms | 542.88353ms | 542.88353ms |
-| propagate: 10000 * 10000 | 5.391814406s | 5.326932802s | 5.379107299s | 5.69781948s | 5.69781948s |
-+--------------------------+--------------+--------------+--------------+-------------+-------------+
```


## Basic usage

```go
import (
	"testing"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/stretchr/testify/assert"
)

// from README
func TestBasics(t *testing.T) {
	rs := alien.CreateReactiveSystem(func(err error) {
		assert.FailNow(t, err.Error())
	})
	count := alien.Signal(rs, 1)
	doubleCount := alien.Computed(rs, func(oldValue int) int {
		return count.Value() * 2
	})

	stopEffect := alien.Effect(rs, func() error {
		return nil
	})
	defer stopEffect()

	assert.Equal(t, 2, doubleCount.Value())

	count.SetValue(2)

	assert.Equal(t, 4, doubleCount.Value())
}
```

## Credits

This is a Dart port of the excellent [stackblitz/alien-signals](https://github.com/stackblitz/alien-signals) library.