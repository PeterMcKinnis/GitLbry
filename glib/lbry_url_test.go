package glib

import (
	"testing"
)

func TestNewLbryUrl(t *testing.T) {
	type args struct {
		x string
	}
	tests := []struct {
		url    string
		wantOk bool
	}{
		{ wantOk: true, url: "lbry://@channel:EF234/stream"},
		{ wantOk: true, url: "lbry://@channel:EF234/stream$1"},

		{ wantOk: true, url: "lbry://@channel"},
		{ wantOk: true, url: "lbry://@channel$1"},
		{ wantOk: true, url: "lbry://@channel:EF234"},
		{ wantOk: true, url: "lbry://@channel*234"},
		{ wantOk: true, url: "lbry://@chaNNel"},

		{ wantOk: true, url: "lbry://stream"},
		{ wantOk: true, url: "lbry://stream$1"},
		{ wantOk: true, url: "lbry://stream:EF234"},
		{ wantOk: true, url: "lbry://stream*234"},
		{ wantOk: true, url: "lbry://stream"},

		{ wantOk: true, url: "lbry://@channel/stream"},
		{ wantOk: true, url: "lbry://@channel/stream$1"},
		{ wantOk: true, url: "lbry://@channel/stream:EF234"},
		{ wantOk: true, url: "lbry://@channel/stream*234"},
		{ wantOk: true, url: "lbry://@channel/stream"},

		{ wantOk: true, url: "lbry://@channel:EF234/stream"},
		{ wantOk: true, url: "lbry://@channel:EF234/stream$1"},
		{ wantOk: true, url: "lbry://@channel:EF234/stream:EF234"},
		{ wantOk: true, url: "lbry://@channel:EF234/stream*234"},
		{ wantOk: true, url: "lbry://@channel:EF234/stream"},

		{ wantOk: false, url: `lbry:\\@channel`},
		{ wantOk: false, url: "@channel"},

	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			ok := IsUrlValid(tt.url)
			if tt.wantOk != ok {
				t.Errorf("IsUrlValid() = %v, want %v", ok, tt.wantOk)
			}
		})
	}
}
