package super

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/djmitche/thespian"
	"github.com/stretchr/testify/require"
)

type server struct {
	serverBase

	padTime time.Duration
}

func (a *server) handleUpper(ctx context.Context, input string) string {
	time.Sleep(a.padTime)
	return strings.ToUpper(input)
}

func (a *server) handleLower(ctx context.Context, input string) string {
	time.Sleep(a.padTime)
	return strings.ToLower(input)
}

func TestBasicRPC(t *testing.T) {
	rt := thespian.NewRuntime()
	server := ServerBuilder{}.spawn(rt)
	defer server.Stop()

	go rt.Run()

	up, err := server.Upper(context.TODO(), "foo")
	require.NoError(t, err)
	require.Equal(t, "FOO", up)

	down, err := server.Lower(context.TODO(), "BAR")
	require.NoError(t, err)
	require.Equal(t, "bar", down)
}

func TestContext(t *testing.T) {
	rt := thespian.NewRuntime()
	server := ServerBuilder{
		server: server{
			padTime: 1 * time.Second,
		},
	}.spawn(rt)
	defer server.Stop()

	go rt.Run()

	ctx, cx := context.WithDeadline(context.Background(), time.Now().Add(10*time.Millisecond))
	defer cx()
	up, err := server.Upper(ctx, "foo")
	require.Equal(t, context.DeadlineExceeded, err)
	require.NotEqual(t, "FOO", up)
}
