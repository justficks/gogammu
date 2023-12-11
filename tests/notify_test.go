package tests

import (
	"reflect"
	"testing"

	gammu "github.com/justficks/gogammu"
)

func TestParseNotification(t *testing.T) {
	tests := []struct {
		input    string
		expected gammu.Notify
		err      bool
	}{
		{
			input: "msg 123123123 5",
			expected: gammu.Notify{
				Type:       "msg",
				PhoneID:    "123123123",
				MessageIDs: []string{"5"},
			},
			err: false,
		},
		{
			input: "msg 123123123 5 6 7",
			expected: gammu.Notify{
				Type:       "msg",
				PhoneID:    "123123123",
				MessageIDs: []string{"5", "6", "7"},
			},
			err: false,
		},
		{
			input: "err 123123123 INIT",
			expected: gammu.Notify{
				Type:      "err",
				PhoneID:   "123123123",
				ErrorType: "INIT",
			},
			err: false,
		},
		{
			input: "call 123123123 from +71234567890",
			expected: gammu.Notify{
				Type:    "call",
				PhoneID: "123123123",
				CallArg: []string{"from", "+71234567890"},
			},
			err: false,
		},
		{
			input: "unknown 123123123",
			err:   true,
		},
	}

	for _, tt := range tests {
		result, err := gammu.ParseNotify(tt.input)
		if err != nil && !tt.err {
			t.Errorf("unexpected error for input %q: %v", tt.input, err)
		}
		if err == nil && tt.err {
			t.Errorf("expected error for input %q, but got none", tt.input)
		}
		if !reflect.DeepEqual(result, tt.expected) {
			t.Errorf("for input %q, expected %+v but got %+v", tt.input, tt.expected, result)
		}
	}
}
