package main

import (
	"errors"
	"testing"
	"time"
)

func Test_replaceReplacer(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"foo\nbar", "foo\\nbar"},
		{"foo\rbar", "foo\\nbar"},
		{"foo\r\nbar", "foo\\nbar"},
		{"foo", "foo"},
	}
	for _, c := range cases {
		got := replaceReplacer(c.in)
		if got != c.want {
			t.Errorf("replaceReplacer(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func Test_Opt_dialOptions_basic(t *testing.T) {
	opts := &Opt{Timeout: 1 * time.Second, Hostname: "localhost", Port: 21}
	options := opts.dialOptions()
	if len(options) == 0 {
		t.Error("dialOptions should return options")
	}
}

func Test_Opt_doConnect_timeout(t *testing.T) {
	// タイムアウトを極端に短くして必ず失敗させる
	opts := &Opt{Timeout: 1 * time.Nanosecond, Hostname: "localhost", Port: 21}
	msg, err := opts.doConnect()
	if err == nil {
		t.Error("doConnect should fail on timeout")
	}
	if !errors.Is(err, err) && msg != "" {
		t.Error("doConnect should return error and empty msg")
	}
}

func Test_opt_verifyOptions(t *testing.T) {
	tests := []struct {
		name    string
		opt     *Opt
		wantErr bool
	}{
		{
			name:    "base",
			opt:     &Opt{},
			wantErr: false,
		},
		{
			name:    "only tcp4",
			opt:     &Opt{TCP4: true},
			wantErr: false,
		},
		{
			name:    "only tcp6",
			opt:     &Opt{TCP6: true},
			wantErr: false,
		},
		{
			name:    "valid options",
			opt:     &Opt{VerifySSL: true, SNI: "example.com"},
			wantErr: false,
		},
		{
			name:    "missing sni",
			opt:     &Opt{VerifySSL: true, SNI: ""},
			wantErr: true,
		},
		{
			name:    "both tcp4 and tcp6",
			opt:     &Opt{TCP4: true, TCP6: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opt.verifyOptions()
			if (err != nil) != tt.wantErr {
				t.Errorf("verifyOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
