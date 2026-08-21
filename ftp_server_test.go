package main

import (
	"crypto/tls"
	"errors"
	"io"
	"log/slog"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

// ftpTestDriver はテスト用の最小 FTP サーバドライバです。
type ftpTestDriver struct {
	tls       bool
	tlsReq    ftpserver.TLSRequirement
	certPEM   []byte
	keyPEM    []byte
	serverDir string
}

func (d *ftpTestDriver) GetSettings() (*ftpserver.Settings, error) {
	return &ftpserver.Settings{
		ListenAddr:          "127.0.0.1:0",
		Banner:              "check_ftp2 test server",
		IdleTimeout:         10,
		ConnectionTimeout:   5,
		DisableActiveMode:   true,
		TLSRequired:         d.tlsReq,
		DefaultTransferType: ftpserver.TransferTypeBinary,
	}, nil
}

func (d *ftpTestDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	return "check_ftp2 test server", nil
}

func (d *ftpTestDriver) ClientDisconnected(cc ftpserver.ClientContext) {}

func (d *ftpTestDriver) AuthUser(_ ftpserver.ClientContext, user, pass string) (ftpserver.ClientDriver, error) {
	if user == "test" && pass == "test" {
		fs := afero.NewBasePathFs(afero.NewOsFs(), d.serverDir)
		return &ftpTestClientDriver{Fs: fs}, nil
	}
	return nil, errors.New("bad username or password")
}

func (d *ftpTestDriver) GetTLSConfig() (*tls.Config, error) {
	if !d.tls {
		return nil, errors.New("TLS is not configured")
	}
	keypair, err := tls.X509KeyPair(d.certPEM, d.keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{keypair},
	}, nil
}

type ftpTestClientDriver struct {
	afero.Fs
}

func newFtpTestDriver(t *testing.T, tls bool, tlsReq ftpserver.TLSRequirement) *ftpTestDriver {
	t.Helper()
	dir, err := os.MkdirTemp("", "check_ftp2_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	d := &ftpTestDriver{
		tls:       tls,
		tlsReq:    tlsReq,
		serverDir: dir,
	}

	if tls {
		certPEM, keyPEM, err := generateSelfSignedCert()
		if err != nil {
			t.Fatal(err)
		}
		d.certPEM = certPEM
		d.keyPEM = keyPEM
	}

	return d
}

func startTestFTPServer(t *testing.T, driver *ftpTestDriver) *ftpserver.FtpServer {
	t.Helper()

	server := ftpserver.NewFtpServer(driver)
	server.Logger = slog.New(slog.NewTextHandler(io.Discard, nil))

	if err := server.Listen(); err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	t.Cleanup(func() {
		if err := server.Stop(); err != nil {
			t.Logf("failed to stop server: %v", err)
		}
	})

	go func() {
		if err := server.Serve(); err != nil && !errors.Is(err, net.ErrClosed) {
			t.Logf("server serve error: %v", err)
		}
	}()

	// サーバが実際に接続を受け付けるまで少し待つ
	deadline := time.Now().Add(2 * time.Second)
	for server.Addr() == "" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	return server
}

func parseServerAddr(t *testing.T, addr string) (string, int) {
	t.Helper()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("failed to split host port: %v", err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("failed to parse port: %v", err)
	}
	return host, port
}

func TestDoConnect_plainFTP(t *testing.T) {
	driver := newFtpTestDriver(t, false, ftpserver.ClearOrEncrypted)
	server := startTestFTPServer(t, driver)

	host, port := parseServerAddr(t, server.Addr())
	opts := &Opt{Timeout: 5 * time.Second, Hostname: host, Port: port}

	msg, err := opts.doConnect()
	if err != nil {
		t.Fatalf("doConnect failed: %v", err)
	}
	if msg == "" {
		t.Fatal("doConnect returned empty message")
	}
	t.Log(msg)
}

func TestDoConnect_implicitTLS(t *testing.T) {
	driver := newFtpTestDriver(t, true, ftpserver.ImplicitEncryption)
	server := startTestFTPServer(t, driver)

	host, port := parseServerAddr(t, server.Addr())
	opts := &Opt{
		Timeout:  5 * time.Second,
		Hostname: host,
		Port:     port,
		SSL:      true,
	}

	msg, err := opts.doConnect()
	if err != nil {
		t.Fatalf("doConnect failed: %v", err)
	}
	if msg == "" {
		t.Fatal("doConnect returned empty message")
	}
	t.Log(msg)
}

func TestDoConnect_explicitTLS(t *testing.T) {
	driver := newFtpTestDriver(t, true, ftpserver.ClearOrEncrypted)
	server := startTestFTPServer(t, driver)

	host, port := parseServerAddr(t, server.Addr())
	opts := &Opt{
		Timeout:  5 * time.Second,
		Hostname: host,
		Port:     port,
		Explicit: true,
	}

	msg, err := opts.doConnect()
	if err != nil {
		t.Fatalf("doConnect failed: %v", err)
	}
	if msg == "" {
		t.Fatal("doConnect returned empty message")
	}
	t.Log(msg)
}
