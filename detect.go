package gammu

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func DetectDevices() (string, error) {
	out, err := exec.Command("gammu-detect", "-b").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gammu-detect error: %v, output: %s", err, string(out))
	}

	err = os.WriteFile("/etc/gammurc", out, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("create /etc/gammurc error: %s", err)
	}

	return string(out), nil
}

func ExtractUSBDevices(s string) map[int]string {
	result := make(map[int]string)

	for _, section := range strings.Split(s, "[gammu") {
		numMatch := regexp.MustCompile(`(\d*)]`).FindStringSubmatch(section)
		if len(numMatch) == 0 {
			continue
		}

		numStr := numMatch[1]
		if numStr == "" {
			numStr = "0"
		}

		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}

		deviceMatch := regexp.MustCompile(`device = (.*)`).FindStringSubmatch(section)
		if len(deviceMatch) > 1 && strings.Contains(deviceMatch[1], "USB") {
			result[num] = strings.TrimSpace(deviceMatch[1])
		}
	}

	return result
}
