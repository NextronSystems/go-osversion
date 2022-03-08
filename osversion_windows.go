//+build windows

package osversion

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func Get() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "Software\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("could not open registry key for CurrentVersion: %s", err)
	}
	defer k.Close()
	productName, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return "", fmt.Errorf("could not get ProductName in CurrentVersion registry key: %s", err)
	}

	if strings.Contains(productName, "Windows 10") { // check build number to determine whether it's actually Windows 11
		buildNumberStr, _, err := k.GetStringValue(`CurrentBuildNumber`)
		if err == nil {
			if buildNumber, err := strconv.Atoi(buildNumberStr); err == nil && buildNumber >= 22000 {
				productName = strings.Replace(productName, "Windows 10", "Windows 11", 1)
			}
		}
	}
	return productName, nil
}
