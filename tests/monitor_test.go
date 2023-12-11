package tests

import (
	"testing"

	gammu "github.com/justficks/gogammu"
)

func TestParseResponseMonitor(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output *gammu.ModemMonitor
	}{
		{
			name:  "Valid input",
			input: "PhoneID: 12345\nIMEI: 98765\nIMSI: 4321\nSent: 100\nReceived: 90\nFailed: 10\nBatterPercent: 80\nNetworkSignal: Strong",
			output: &gammu.ModemMonitor{
				PhoneID:       "12345",
				IMEI:          "98765",
				IMSI:          "4321",
				Sent:          "100",
				Received:      "90",
				Failed:        "10",
				BatterPercent: "80",
				NetworkSignal: "Strong",
			},
		},
		{
			name:  "Extra spaces",
			input: " PhoneID : 12345 \n IMEI: 98765 ",
			output: &gammu.ModemMonitor{
				PhoneID: "12345",
				IMEI:    "98765",
			},
		},
		{
			name:   "Empty input",
			input:  "",
			output: &gammu.ModemMonitor{},
		},
		{
			name:   "Invalid line format",
			input:  "PhoneID12345\nIMEI 98765",
			output: &gammu.ModemMonitor{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gammu.ParseMonitor(tt.input)
			if !isEqual(got, tt.output) {
				t.Errorf("parseResponseMonitor() = %v, want %v", got, tt.output)
			}
		})
	}
}

func isEqual(a, b *gammu.ModemMonitor) bool {
	return a.PhoneID == b.PhoneID &&
		a.IMEI == b.IMEI &&
		a.IMSI == b.IMSI &&
		a.Sent == b.Sent &&
		a.Received == b.Received &&
		a.Failed == b.Failed &&
		a.BatterPercent == b.BatterPercent &&
		a.NetworkSignal == b.NetworkSignal
}
