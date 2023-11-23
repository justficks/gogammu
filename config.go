package gammu

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (g *Gammu) CreateConfig(modem *Modem) (string, string, error) {
	configContent := g.ConfigContent(modem.Device, modem.IMSI)

	cfgFilePath := filepath.Join(g.CfgDir, modem.IMSI)
	err := os.WriteFile(cfgFilePath, []byte(configContent), os.ModePerm)
	if err != nil {
		return configContent, cfgFilePath, fmt.Errorf("create %s error %s", cfgFilePath, err)
	}

	return configContent, cfgFilePath, nil
}

func (g *Gammu) ConfigContent(device, modemId string) string {
	parts := []string{
		"[gammu]",
		"device = " + device,
		"connection = at",
		"",
		"[smsd]",
		"Service = sql",
		"Driver = native_pgsql",
		"Host = " + g.DbHost,
		"User = " + g.DbUser,
		"Password = " + g.DbPass,
		"Database = smsd",
		"",
		"LogFile = " + filepath.Join(g.LogDir, modemId),
		"PhoneId = " + modemId,
		"MultipartTimeout = 120",
		"RunOnReceive = " + g.Script + " msg " + modemId,
		"RunOnFailure = " + g.Script + " err " + modemId,
		"RunOnCall = " + g.Script + " call " + modemId,
	}
	return strings.Join(parts, "\n")
}

type Notify struct {
	Type       string
	PhoneID    string // IMSI of SIM card
	MessageIDs []string
	ErrorType  string
	CallArg    []string
}

func ParseNotify(input string) (*Notify, error) {
	parts := strings.Fields(input)

	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid input: %s", input)
	}

	notificationType := parts[0]
	phoneID := parts[1]

	switch notificationType {
	case "msg":
		return &Notify{
			Type:       notificationType,
			PhoneID:    phoneID,
			MessageIDs: parts[2:],
		}, nil
	case "err":
		return &Notify{
			Type:      notificationType,
			PhoneID:   phoneID,
			ErrorType: parts[2],
		}, nil
	case "call":
		return &Notify{
			Type:    notificationType,
			PhoneID: phoneID,
			CallArg: parts[2:],
		}, nil
	default:
		return nil, fmt.Errorf("unknown notification type: %s", notificationType)
	}
}

func CreateNotifyScript(scriptPath string) error {
	parts := []string{
		"#!/bin/bash",
		"args=\"$@\"",
		"curl -d \"$args\" http://localhost:3000/notify",
	}
	content := strings.Join(parts, "\n")

	err := os.WriteFile(scriptPath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}

	out, err := exec.Command("chmod", "+x", scriptPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("chmod +x %s error: %s, output: %s", scriptPath, err, string(out))
	}

	return nil
}
