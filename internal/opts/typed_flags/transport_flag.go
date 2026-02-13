package typed_flags

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Transport string

const (
	TransportStdio Transport = "stdio"
	TransportHTTP  Transport = "http"
)

var TransportValues = []Transport{
	TransportStdio,
	TransportHTTP,
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ flags.Completer   = (*Transport)(nil)
	_ flags.Unmarshaler = (*Transport)(nil)
)

func (t *Transport) Complete(match string) (completions []flags.Completion) {
	for _, v := range TransportValues {
		val := string(v)
		if match == "" || strings.HasPrefix(val, strings.ToLower(match)) {
			completions = append(completions, flags.Completion{
				Item:        val,
				Description: "",
			})
		}
	}
	return
}

// UnmarshalFlag validates the value is one of the allowed values.
func (t *Transport) UnmarshalFlag(value string) error {
	for _, v := range TransportValues {
		if string(v) == value {
			*t = v
			return nil
		}
	}
	return fmt.Errorf("invalid transport: %s (valid: %v)", value, TransportValues)
}
