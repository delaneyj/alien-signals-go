# Alien Signals for Go


<p align="center">
  <img src="mascot.png" width="384"><br>
<p>

The lightest signal library for Go, ported from [stackblitz/alien-signals](https://github.com/stackblitz/alien-signals).
[![Go Reference](https://pkg.go.dev/badge/github.com/delaneyj/alien-signals-go.svg)](https://pkg.go.dev/github.com/delaneyj/alien-signals-go)

> [!TIP]
> `alien_signals` is the fastest signal library currently.

## Benchmarks against original TypeScript implementation
Node with JIT

```md
| benchmark                | avg                                     | min         | p75         | p99         | max         |
| ------------------------ | --------------------------------------- | ----------- | ----------- | ----------- | ----------- |
| propagate: 1 * 1         | ` 48.61 ns/iter`                        | ` 44.73 ns` | ` 47.97 ns` | ` 73.81 ns` | `120.40 ns` |
| propagate: 1 * 10        | `268.84 ns/iter`                        | `256.60 ns` | `278.34 ns` | `294.42 ns` | `506.19 ns` |
| propagate: 1 * 100       | `  2.93 µs/iter`                        | `  2.90 µs` | `  2.93 µs` | `  2.96 µs` | `  2.96 µs` |
| propagate: 1 * 1000      | ` 34.32 µs/iter`                        | ` 34.10 µs` | ` 34.46 µs` | ` 34.47 µs` | ` 34.56 µs` |
| propagate: 1 * 10000     | error: Maximum call stack size exceeded |
| propagate: 10 * 1        | `437.01 ns/iter`                        | `412.83 ns` | `445.33 ns` | `477.34 ns` | `577.09 ns` |
| propagate: 10 * 10       | `  3.04 µs/iter`                        | `  3.01 µs` | `  3.04 µs` | `  3.12 µs` | `  3.29 µs` |
| propagate: 10 * 100      | ` 29.11 µs/iter`                        | ` 28.85 µs` | ` 29.34 µs` | ` 29.49 µs` | ` 29.52 µs` |
| propagate: 10 * 1000     | `296.37 µs/iter`                        | `272.57 µs` | `295.65 µs` | `393.12 µs` | `510.07 µs` |
| propagate: 10 * 10000    | error: Maximum call stack size exceeded |
| propagate: 100 * 1       | `  4.65 µs/iter`                        | `  4.60 µs` | `  4.66 µs` | `  4.71 µs` | `  4.77 µs` |
| propagate: 100 * 10      | ` 33.41 µs/iter`                        | ` 33.23 µs` | ` 33.54 µs` | ` 33.66 µs` | ` 33.67 µs` |
| propagate: 100 * 100     | `305.75 µs/iter`                        | `285.65 µs` | `306.11 µs` | `389.72 µs` | `422.69 µs` |
| propagate: 100 * 1000    | `  4.77 ms/iter`                        | `  4.56 ms` | `  4.84 ms` | `  5.04 ms` | `  5.11 ms` |
| propagate: 100 * 10000   | error: Maximum call stack size exceeded |
| propagate: 1000 * 1      | ` 45.51 µs/iter`                        | ` 45.22 µs` | ` 45.57 µs` | ` 45.74 µs` | ` 46.36 µs` |
| propagate: 1000 * 10     | `359.52 µs/iter`                        | `332.85 µs` | `362.54 µs` | `463.21 µs` | `558.22 µs` |
| propagate: 1000 * 100    | `  5.51 ms/iter`                        | `  5.28 ms` | `  5.57 ms` | `  5.74 ms` | `  5.85 ms` |
| propagate: 1000 * 1000   | ` 57.36 ms/iter`                        | ` 56.54 ms` | ` 57.25 ms` | ` 57.46 ms` | ` 61.10 ms` |
| propagate: 1000 * 10000  | error: Maximum call stack size exceeded |
| propagate: 10000 * 1     | `485.35 µs/iter`                        | `448.03 µs` | `487.39 µs` | `625.87 µs` | `671.36 µs` |
| propagate: 10000 * 10    | `  9.25 ms/iter`                        | `  8.87 ms` | `  9.37 ms` | `  9.78 ms` | ` 10.03 ms` |
| propagate: 10000 * 100   | ` 62.25 ms/iter`                        | ` 61.21 ms` | ` 61.86 ms` | ` 63.15 ms` | ` 66.95 ms` |
| propagate: 10000 * 1000  | `676.41 ms/iter`                        | `659.72 ms` | `675.46 ms` | `710.27 ms` | `734.99 ms` |
| propagate: 10000 * 10000 | error: Maximum call stack size exceeded |
```

Go
```md
+--------------------------+--------------+--------------+--------------+--------------+--------------+
| BENCHMARK                |          AVG |          MIN |          P75 |          P99 |          MAX |
+--------------------------+--------------+--------------+--------------+--------------+--------------+
| propagate: 1 * 1         |        125ns |        110ns |        111ns |        210ns |      1.333µs |
| propagate: 1 * 10        |        688ns |        581ns |        712ns |        872ns |      1.032µs |
| propagate: 1 * 100       |      3.915µs |      3.767µs |      3.807µs |      4.709µs |       5.32µs |
| propagate: 1 * 1000      |     37.467µs |     36.741µs |     37.151µs |     45.127µs |     58.773µs |
| propagate: 1 * 10000     |    389.447µs |    361.536µs |      378.9µs |    514.461µs |    1.54186ms |
| propagate: 10 * 1        |        675ns |        661ns |        672ns |        762ns |        972ns |
| propagate: 10 * 10       |      4.332µs |      4.298µs |      4.329µs |      4.579µs |      4.709µs |
| propagate: 10 * 100      |     39.324µs |     38.945µs |     39.406µs |     43.173µs |     43.784µs |
| propagate: 10 * 1000     |    393.467µs |    373.399µs |    391.875µs |    580.558µs |    807.164µs |
| propagate: 10 * 10000    |   3.986958ms |   3.827672ms |   4.016606ms |   4.407768ms |   4.799042ms |
| propagate: 100 * 1       |      6.282µs |      6.051µs |      6.543µs |      7.374µs |      7.504µs |
| propagate: 100 * 10      |     40.828µs |     40.367µs |     40.609µs |     47.932µs |     56.579µs |
| propagate: 100 * 100     |    384.984µs |    378.459µs |    386.725µs |    426.411µs |    468.873µs |
| propagate: 100 * 1000    |   4.087396ms |   3.907364ms |   4.046493ms |   4.881872ms |    5.47789ms |
| propagate: 100 * 10000   |  43.243626ms |  42.381882ms |  43.421143ms |  48.683918ms |  49.529747ms |
| propagate: 1000 * 1      |     60.236µs |     59.034µs |     60.396µs |     69.644µs |     73.842µs |
| propagate: 1000 * 10     |    439.561µs |    416.562µs |     446.59µs |    463.663µs |    465.697µs |
| propagate: 1000 * 100    |   4.502427ms |   4.305091ms |   4.542128ms |   4.754507ms |   4.914113ms |
| propagate: 1000 * 1000   |  44.229612ms |  42.923915ms |  45.174851ms |  46.030368ms |  46.405901ms |
| propagate: 1000 * 10000  | 434.184596ms | 426.674225ms | 435.859695ms |  460.31353ms |  462.34876ms |
| propagate: 10000 * 1     |    642.171µs |    610.316µs |    637.768µs |    810.491µs |   1.059511ms |
| propagate: 10000 * 10    |   8.338133ms |   7.336088ms |   8.622125ms |   9.982575ms |  10.264829ms |
| propagate: 10000 * 100   |  52.092456ms |  50.629646ms |  52.723659ms |  56.869794ms |  57.060861ms |
| propagate: 10000 * 1000  | 439.609865ms | 435.477779ms | 441.222123ms | 448.839332ms | 450.429014ms |
| propagate: 10000 * 10000 | 4.530244316s | 4.286658058s | 4.529553663s | 5.666950049s | 6.713975996s |
+--------------------------+--------------+--------------+--------------+--------------+--------------+
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

This is a Go port of the excellent [stackblitz/alien-signals](https://github.com/stackblitz/alien-signals) library.