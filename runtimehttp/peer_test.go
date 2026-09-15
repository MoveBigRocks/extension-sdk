package runtimehttp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/movebigrocks/extension-sdk/runtimeproto"
)

// tempSocketPath returns a socket path short enough for the unix domain socket
// address limit, which the per-test directory names t.TempDir builds exceed on
// macOS.
func tempSocketPath(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "mbr-runtime")
	if err != nil {
		t.Fatalf("create socket dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	return filepath.Join(dir, "runtime.sock")
}

func TestListenPrivateUnixCreatesOwnerOnlySocket(t *testing.T) {
	socketPath := tempSocketPath(t)

	listener, err := listenPrivateUnix(socketPath)
	if err != nil {
		t.Fatalf("listen private unix: %v", err)
	}
	defer listener.Close()

	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("stat socket: %v", err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		t.Fatalf("expected a socket, got mode %v", info.Mode())
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected socket mode 0600, got %04o", perm)
	}
}

func TestNewRuntimeServerServesVerifiedPeer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ForwardedContextMiddleware())
	engine.GET("/who", func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString("user_id"))
	})

	socketPath := tempSocketPath(t)
	server, listener, err := newRuntimeServer(engine, socketPath)
	if err != nil {
		t.Fatalf("new runtime server: %v", err)
	}
	go func() {
		_ = server.Serve(listener)
	}()
	t.Cleanup(func() {
		_ = server.Close()
	})

	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("stat socket: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected socket mode 0600, got %04o", perm)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				var dialer net.Dialer
				return dialer.DialContext(ctx, "unix", socketPath)
			},
		},
	}
	req, err := http.NewRequest(http.MethodGet, "http://unix/who", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set(runtimeproto.HeaderUserID, "usr_123")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("dial runtime socket: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for a same-uid peer, got %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if string(body) != "usr_123" {
		t.Fatalf("expected the forwarded user id to be trusted, got %q", string(body))
	}
}

func TestPeerUIDReadsKernelCredentials(t *testing.T) {
	socketPath := tempSocketPath(t)
	listener, err := listenPrivateUnix(socketPath)
	if err != nil {
		t.Fatalf("listen private unix: %v", err)
	}
	defer listener.Close()

	go func() {
		client, err := net.Dial("unix", socketPath)
		if err == nil {
			defer client.Close()
			time.Sleep(50 * time.Millisecond)
		}
	}()

	conn, err := listener.Accept()
	if err != nil {
		t.Fatalf("accept: %v", err)
	}
	defer conn.Close()

	uid, err := peerUID(conn)
	if err != nil {
		t.Fatalf("read peer uid: %v", err)
	}
	if uid != uint32(os.Geteuid()) {
		t.Fatalf("expected peer uid %d, got %d", os.Geteuid(), uid)
	}
}

func TestAuthorizePeerUIDAdmitsOnlyTheRuntimeAccountAndRoot(t *testing.T) {
	if err := authorizePeerUID(uint32(os.Geteuid())); err != nil {
		t.Fatalf("expected the runtime's own uid to be authorized: %v", err)
	}
	if err := authorizePeerUID(0); err != nil {
		t.Fatalf("expected root to be authorized: %v", err)
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root: every peer uid is the runtime uid or root, so nothing is left to refuse")
	}
	if err := authorizePeerUID(uint32(os.Geteuid()) + 1); err == nil {
		t.Fatal("expected a peer running as another account to be refused")
	}
}

func TestVerifiedPeerListenerRefusesForeignUID(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: every local peer is authorized, so the rejection path cannot be observed")
	}

	socketPath := tempSocketPath(t)
	base, err := listenPrivateUnix(socketPath)
	if err != nil {
		t.Fatalf("listen private unix: %v", err)
	}
	listener := newVerifiedPeerListener(base)
	defer listener.Close()

	// A test process cannot run a second account without privileges, so the
	// kernel reader is swapped for one reporting a foreign uid. What the
	// listener then does with that uid is the production path.
	listener.readPeerUID = func(net.Conn) (uint32, error) {
		return uint32(os.Geteuid()) + 1, nil
	}

	accepted := make(chan net.Conn, 1)
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			accepted <- conn
		}
	}()

	client, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer client.Close()

	if err := client.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	if _, err := client.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Fatalf("expected the listener to close a foreign-uid peer, got %v", err)
	}
	select {
	case <-accepted:
		t.Fatal("a foreign-uid connection was handed to the server")
	default:
	}
}

func TestPeerConnContextMarksOnlyAuthenticatedConnections(t *testing.T) {
	if PeerVerified(peerConnContext(context.Background(), &net.UnixConn{})) {
		t.Fatal("expected a connection that skipped the credential check to stay unverified")
	}
	if !PeerVerified(peerConnContext(context.Background(), &verifiedConn{})) {
		t.Fatal("expected an authenticated connection to be marked verified")
	}
}
