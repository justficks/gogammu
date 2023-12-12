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
		"RunOnReceive = " + g.RunOnMsgScript + modemId,
		"RunOnFailure = " + g.RunOnErrScript + modemId,
		"RunOnCall = " + g.RunOnCallScript + modemId,
	}
	return strings.Join(parts, "\n")
}

type RunOnErrBody struct {
	PhoneID  string // IMSI of SIM card
	ErrorArg []string
}

func ParseRunOnErrBody(input string) (*RunOnErrBody, error) {
	parts := strings.Fields(input)

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid input: %s", input)
	}

	return &RunOnErrBody{
		PhoneID:  parts[0],
		ErrorArg: parts[1:],
	}, nil
}

type RunOnMsgBody struct {
	PhoneID    string // IMSI of SIM card
	MessageIDs []string
}

func ParseRunOnMsgBody(input string) (*RunOnMsgBody, error) {
	parts := strings.Fields(input)

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid input: %s", input)
	}

	return &RunOnMsgBody{
		PhoneID:    parts[0],
		MessageIDs: parts[1:],
	}, nil
}

func CreateRunOnScript(scriptPath, appHttpPort, notifyType string) error {
	parts := []string{
		"#!/bin/bash",
		"args=\"$@\"",
		"curl -d \"$args\" http://localhost:" + appHttpPort + "/run-on/" + notifyType,
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
