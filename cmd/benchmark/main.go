package main

import (
	"fmt"
	"log"
	"os"
	"time"

	alien "github.com/delaneyj/alien-signals-go"
	"github.com/jamiealquiza/tachymeter"
	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {

	ww := []int{1, 10, 100, 1000}
	hh := []int{1, 10, 100, 1000}

	getValue := func(x any) int {
		switch x := x.(type) {
		case *alien.WriteableSignal[int]:
			return x.Value() + 1
		case *alien.ReadonlySignal[int]:
			return x.Value() + 1
		default:
			panic("unknown type")
		}
	}

	tbl := table.NewWriter()
	tbl.SetOutputMirror(os.Stdout)
	tbl.AppendHeader(table.Row{"benchmark", "avg", "min", "p75", "p99", "max"})

	for _, w := range ww {
		for _, h := range hh {
			iters := 10
			tach := tachymeter.New(&tachymeter.Config{Size: iters})

			// fmt.Sprintf("propagate: %dx%d", w, h), func(b *testing.B) {
			rs := alien.CreateReactiveSystem(func(err error) {
				log.Panic(err)
			})
			src := alien.Signal(rs, 1)
			for i := 0; i < w; i++ {
				var last any
				last = src
				for j := 0; j < h; j++ {
					prev := last
					last = alien.Computed(rs, func(oldValue int) int {
						return getValue(prev)
					})
				}

				alien.Effect(rs, func() error {
					getValue(last)
					return nil
				})

			}

			for i := 0; i < iters; i++ {
				start := time.Now()
				src.SetValue(src.Value() + 1)
				tach.AddTime(time.Since(start))
			}

			calc := tach.Calc()
			tbl.AppendRows([]table.Row{
				{
					fmt.Sprintf("propagate: %d * %d", w, h),
					calc.Time.Avg,
					calc.Time.Min,
					calc.Time.P75,
					calc.Time.P99,
					calc.Time.Max,
				},
			})
		}
	}

	tbl.Render()
}
