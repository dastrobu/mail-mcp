package typed_flags

import (
	"testing"

	"github.com/jessevdk/go-flags"
)

func TestTransport_UnmarshalFlag(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    Transport
		wantErr bool
	}{
		{
			name:    "valid stdio",
			value:   "stdio",
			want:    TransportStdio,
			wantErr: false,
		},
		{
			name:    "valid http",
			value:   "http",
			want:    TransportHTTP,
			wantErr: false,
		},
		{
			name:    "invalid transport",
			value:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			value:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "case sensitive",
			value:   "HTTP",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var transport Transport
			err := transport.UnmarshalFlag(tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalFlag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && transport != tt.want {
				t.Errorf("UnmarshalFlag() got = %v, want %v", transport, tt.want)
			}
		})
	}
}

func TestTransport_Complete(t *testing.T) {
	tests := []struct {
		name      string
		match     string
		wantCount int
		wantItems []string
	}{
		{
			name:      "empty match returns all",
			match:     "",
			wantCount: 2,
			wantItems: []string{"stdio", "http"},
		},
		{
			name:      "match stdio",
			match:     "s",
			wantCount: 1,
			wantItems: []string{"stdio"},
		},
		{
			name:      "match http",
			match:     "h",
			wantCount: 1,
			wantItems: []string{"http"},
		},
		{
			name:      "no match",
			match:     "xyz",
			wantCount: 0,
			wantItems: []string{},
		},
		{
			name:      "partial match stdio",
			match:     "std",
			wantCount: 1,
			wantItems: []string{"stdio"},
		},
		{
			name:      "partial match http",
			match:     "ht",
			wantCount: 1,
			wantItems: []string{"http"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var transport Transport
			completions := transport.Complete(tt.match)

			if len(completions) != tt.wantCount {
				t.Errorf("Complete() returned %d completions, want %d", len(completions), tt.wantCount)
			}

			for i, want := range tt.wantItems {
				if i >= len(completions) {
					t.Errorf("Missing completion item: %s", want)
					continue
				}
				if completions[i].Item != want {
					t.Errorf("Complete()[%d].Item = %v, want %v", i, completions[i].Item, want)
				}
			}
		})
	}
}

func TestTransport_ImplementsInterfaces(t *testing.T) {
	var transport Transport

	// Check Completer interface
	_, ok := interface{}(&transport).(flags.Completer)
	if !ok {
		t.Error("Transport does not implement flags.Completer interface")
	}

	// Check Unmarshaler interface
	_, ok = interface{}(&transport).(flags.Unmarshaler)
	if !ok {
		t.Error("Transport does not implement flags.Unmarshaler interface")
	}
}

func TestTransportValues(t *testing.T) {
	if len(TransportValues) != 2 {
		t.Errorf("TransportValues should have 2 values, got %d", len(TransportValues))
	}

	expectedValues := map[Transport]bool{
		TransportStdio: true,
		TransportHTTP:  true,
	}

	for _, v := range TransportValues {
		if !expectedValues[v] {
			t.Errorf("Unexpected value in TransportValues: %v", v)
		}
		delete(expectedValues, v)
	}

	if len(expectedValues) > 0 {
		t.Errorf("Missing values in TransportValues: %v", expectedValues)
	}
}
