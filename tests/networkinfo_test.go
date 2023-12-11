package tests

import (
	"reflect"
	"testing"

	gammu "github.com/justficks/gogammu"
)

// Тесты для функции parseGammuNetworkInfoResponse
func TestParseGammuNetworkInfoResponse(t *testing.T) {
	tests := []struct {
		name string
		data string
		want *gammu.ModemNetwork
	}{
		{
			name: "Test 1",
			data: `Network state        : registration to network denied
			Packet network state : requesting network
			GPRS                 : detached`,
			want: &gammu.ModemNetwork{
				NetworkState:       "registration to network denied",
				PacketNetworkState: "requesting network",
				GPRS:               "detached",
			},
		},
		{
			name: "Test 2",
			data: `Network state        : home network
			Network              : 250 20 (Tele2, Russian Federation), LAC B32A, CID 06A7
			Name in phone        : "MOTIV"
			Packet network state : not logged into network
			GPRS                 : detached`,
			want: &gammu.ModemNetwork{
				NetworkState:       "home network",
				Network:            "250 20 (Tele2, Russian Federation), LAC B32A, CID 06A7",
				NameInPhone:        "\"MOTIV\"",
				PacketNetworkState: "not logged into network",
				GPRS:               "detached",
			},
		},
		{
			name: "Test 3",
			data: `Network state        : home network
			Network              : 250 20 (Tele2, Russian Federation), LAC B32A, CID 06A7
			Name in phone        : "MOTIV"
			Packet network state : home network
			Packet network       : 250 20 (Tele2, Russian Federation), LAC B32A, CID 06A7
			Name in phone        : "MOTIV"
			GPRS                 : attached`,
			want: &gammu.ModemNetwork{
				NetworkState:       "home network",
				Network:            "250 20 (Tele2, Russian Federation), LAC B32A, CID 06A7",
				NameInPhone:        "\"MOTIV\"",
				PacketNetworkState: "home network",
				PacketNetwork:      "250 20 (Tele2, Russian Federation), LAC B32A, CID 06A7",
				GPRS:               "attached",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gammu.ParseNetwork(tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
