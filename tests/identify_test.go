package tests

import (
	"testing"

	gammu "github.com/justficks/gogammu"
)

func TestParseIdentifyResponse(t *testing.T) {
	tests := []struct {
		input    string
		expected *gammu.ModemIdentify
	}{
		{
			input: "Device: MyDevice\nManufacturer: MyManufacturer\nModel: MyModel\nFirmware: MyFirmware\nIMEI: 123456789012345\nSIM IMSI: 123456789012345",
			expected: &gammu.ModemIdentify{
				Device:       "MyDevice",
				Manufacturer: "MyManufacturer",
				Model:        "MyModel",
				Firmware:     "MyFirmware",
				IMEI:         "123456789012345",
				IMSI:         "123456789012345",
			},
		},
		{
			input: "Device: Device2\nManufacturer: Manufacturer2\nModel: Model2\nFirmware: Firmware2\nIMEI: 234567890123456\nSIM IMSI: 234567890123456",
			expected: &gammu.ModemIdentify{
				Device:       "Device2",
				Manufacturer: "Manufacturer2",
				Model:        "Model2",
				Firmware:     "Firmware2",
				IMEI:         "234567890123456",
				IMSI:         "234567890123456",
			},
		},
		{
			input: "Device: Device3\nManufacturer: Manufacturer3\nModel:\nFirmware:\nIMEI:\nSIM IMSI: 345678901234567",
			expected: &gammu.ModemIdentify{
				Device:       "Device3",
				Manufacturer: "Manufacturer3",
				Model:        "",
				Firmware:     "",
				IMEI:         "",
				IMSI:         "345678901234567",
			},
		},
	}

	for _, tt := range tests {
		result := gammu.ParseIdentify(tt.input)
		if *result != *tt.expected {
			t.Errorf("Expected:\n %+v but got:\n %+v for input:\n%s", tt.expected, result, tt.input)
		}
	}
}
