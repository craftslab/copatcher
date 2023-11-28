package buildkit

import (
	"context"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	controlapi "github.com/moby/buildkit/api/services/control"
	types "github.com/moby/buildkit/api/types"
	"github.com/moby/buildkit/client"
	gateway "github.com/moby/buildkit/frontend/gateway/pb"
	"github.com/moby/buildkit/util/apicaps"
	caps "github.com/moby/buildkit/util/apicaps/pb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	prefix = "unix://"
)

type mockControlServer struct {
	controlapi.ControlServer
}

func (s *mockControlServer) ListWorkers(context.Context, *controlapi.ListWorkersRequest) (*controlapi.ListWorkersResponse, error) {
	return &controlapi.ListWorkersResponse{
		Record: []*types.WorkerRecord{},
	}, nil
}

func (s *mockControlServer) Session(controlapi.Control_SessionServer) error {
	return nil
}

func (s *mockControlServer) Status(*controlapi.StatusRequest, controlapi.Control_StatusServer) error {
	return nil
}

func (s *mockControlServer) Solve(context.Context, *controlapi.SolveRequest) (*controlapi.SolveResponse, error) {
	return &controlapi.SolveResponse{}, nil
}

type mockLLBBridgeServer struct {
	gateway.LLBBridgeServer
	caps []caps.APICap
}

func (m *mockLLBBridgeServer) Ping(context.Context, *gateway.PingRequest) (*gateway.PongResponse, error) {
	return &gateway.PongResponse{
		FrontendAPICaps: m.caps,
		LLBCaps:         m.caps,
	}, nil
}

func (m *mockLLBBridgeServer) Solve(context.Context, *gateway.SolveRequest) (*gateway.SolveResponse, error) {
	return &gateway.SolveResponse{}, nil
}

func makeCapList(capIDs ...apicaps.CapID) []caps.APICap {
	var (
		ls apicaps.CapList
		c  = make([]apicaps.Cap, 0, len(capIDs))
	)

	for _, id := range capIDs {
		c = append(c, apicaps.Cap{
			ID:      id,
			Enabled: true,
		})
	}

	ls.Init(c...)

	return ls.All()
}

func newMockBuildkitAPI(t *testing.T, c ...apicaps.CapID) string {
	tmp := t.TempDir()

	l, err := net.Listen("unix", filepath.Join(tmp, "buildkitd.sock"))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = l.Close()
	})

	srv := grpc.NewServer()
	t.Cleanup(srv.Stop)

	capList := makeCapList(c...)
	gateway.RegisterLLBBridgeServer(srv, &mockLLBBridgeServer{
		LLBBridgeServer: &gateway.UnimplementedLLBBridgeServer{},
		caps:            capList,
	})

	go func() {
		_ = srv.Serve(l)
	}()

	control := &mockControlServer{
		ControlServer: &controlapi.UnimplementedControlServer{},
	}

	controlapi.RegisterControlServer(srv, control)

	return l.Addr().String()
}

func unwrapErrors(err error) []error {
	// `errors.Unwrap` uses this interface
	// buildkit errors may be wrapped in this
	type stdUnwrap interface {
		Unwrap() error
	}

	// The type used by `errors.Join` uses this interface
	type joinedUnwrap interface {
		Unwrap() []error
	}

	var out []error

	switch v := err.(type) {
	case stdUnwrap:
		return unwrapErrors(v.Unwrap())
	case joinedUnwrap:
		for _, e := range v.Unwrap() {
			// multiple calls to `errors.Join` may result in nested wraps, so recurse on those errors
			out = append(out, unwrapErrors(e)...)
		}
	default:
		out = append(out, err)
	}

	return out
}

func checkMissingCapsError(t *testing.T, err error, c ...apicaps.CapID) {
	t.Helper()
	lsErr := unwrapErrors(err)
	found := make(map[apicaps.CapID]bool, len(c))

	for _, err := range lsErr {
		check := &apicaps.CapError{}
		if errors.As(err, &check) {
			found[check.ID] = true
		}
	}

	if len(found) != len(c) {
		t.Errorf("expected %d errors, got: %d", len(c), len(found))
		t.Error(lsErr)
	}
}

func TestGetServerNameFromAddr(t *testing.T) {
	testCases := []struct {
		name string
		addr string
		want string
	}{
		{
			name: "hostname",
			addr: "tcp://hostname:1234",
			want: "hostname",
		},
		{
			name: "IP address",
			addr: "tcp://127.0.0.1:1234",
			want: "127.0.0.1",
		},
		{
			name: "invalid URL",
			addr: "hostname:1234",
			want: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := getServerNameFromAddr(tc.addr)
			if got != tc.want {
				t.Errorf("getServerNameFromAddr(%q) = %q, want %q", tc.addr, got, tc.want)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()

	t.Run("custom buildkit addr", func(t *testing.T) {
		t.Run("missing caps", func(t *testing.T) {
			t.Parallel()
			addr := newMockBuildkitAPI(t)
			ctxT, cancel := context.WithTimeout(ctx, time.Second)
			bkOpts := Opts{
				Addr: prefix + addr,
			}
			_client, err := NewClient(ctxT, bkOpts)
			cancel()
			assert.NoError(t, err)
			defer func(c *client.Client) {
				_ = c.Close()
			}(_client)
			ctxT, cancel = context.WithTimeout(ctx, time.Second)
			err = ValidateClient(ctxT, _client)
			cancel()
			checkMissingCapsError(t, err, requiredCaps...)
		})

		t.Run("with caps", func(t *testing.T) {
			t.Parallel()
			addr := newMockBuildkitAPI(t, requiredCaps...)

			ctxT, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			bkOpts := Opts{
				Addr: prefix + addr,
			}
			_client, err := NewClient(ctxT, bkOpts)
			assert.NoError(t, err)
			defer func(c *client.Client) {
				_ = c.Close()
			}(_client)
			err = ValidateClient(ctxT, _client)
			assert.NoError(t, err)
		})
	})
}
