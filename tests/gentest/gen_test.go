package gentest

import (
	"testing"
	"time"

	"github.com/djmitche/thespian"
	"github.com/stretchr/testify/require"
)

func TestConcatenator(t *testing.T) {
	rt := thespian.NewRuntime()
	conc := NewConcatenator(rt)

	var res string
	go func() {
		conc.Input("abc")
		conc.Input("def")
		conc.Input("ghi")

		time.Sleep(10 * time.Millisecond) // XXX !!!

		ch := make(chan string)
		conc.Output(ch)
		res = <-ch

		conc.Stop()
	}()

	rt.Run()

	require.Equal(t, "abcdefghi", res)
}
