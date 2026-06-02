//go:build linux
// +build linux

package osversion

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	osReleasePath     = "/etc/os-release"
	debianVersionPath = "/etc/debian_version"
	redhatReleasePath = "/etc/redhat-release"
	suseReleasePath   = "/etc/SuSe-release"
)

func Get() (string, error) {
	if version := getFromOSRelease(); version != "" {
		return version, nil
	}
	if version := getFromLSB(); version != "" {
		return version, nil
	}
	if version := getFromDebianVersion(); version != "" {
		return version, nil
	}
	if version := getFromRedhatRelease(); version != "" {
		return version, nil
	}
	if version := getFromSuSeRelease(); version != "" {
		return version, nil
	}
	return getFromUname()
}

func getFromOSRelease() string {
	b, err := readFileSafe(osReleasePath)
	if err != nil {
		return ""
	}
	r := bufio.NewReader(bytes.NewReader(b))
	for {
		line, err := r.ReadString('\n')
		line = strings.TrimRight(line, "\n")
		// PRETTY_NAME="Debian GNU/Linux 9 (stretch)"
		if strings.HasPrefix(line, `PRETTY_NAME="`) && strings.HasSuffix(line, `"`) && len(line) >= 14 {
			return line[13 : len(line)-1]
		}
		if err != nil {
			return ""
		}
	}
}

func getFromDebianVersion() string {
	b, err := readFileSafe(debianVersionPath)
	if err != nil {
		return ""
	}
	r := bufio.NewReader(bytes.NewReader(b))
	line, _ := r.ReadString('\n')
	line = strings.TrimSuffix(line, "\n")
	if line == "" {
		return ""
	}
	return "Debian " + line
}

func getFromRedhatRelease() string {
	b, err := readFileSafe(redhatReleasePath)
	if err != nil {
		return ""
	}
	r := bufio.NewReader(bytes.NewReader(b))
	line, _ := r.ReadString('\n')
	return strings.TrimSuffix(line, "\n")
}

func getFromSuSeRelease() string {
	b, err := readFileSafe(suseReleasePath)
	if err != nil {
		return ""
	}
	r := bufio.NewReader(bytes.NewReader(b))
	line, _ := r.ReadString('\n')
	return strings.TrimSuffix(line, "\n")
}

func getFromLSB() string {
	oscmd := exec.Command("lsb_release", "-si")
	vercmd := exec.Command("lsb_release", "-sr")
	os, err := oscmd.Output()
	if err != nil {
		return ""
	}
	ver, err := vercmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(os)) + " " + strings.TrimSpace(string(ver))
}

func getFromUname() (string, error) {
	oscmd := exec.Command("uname", "-s")
	vercmd := exec.Command("uname", "-r")
	os, err := oscmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not execute uname: %s", err)
	}
	ver, err := vercmd.Output()
	if err != nil {
		return "", fmt.Errorf("could not execute uname: %s", err)
	}
	return strings.TrimSpace(string(os)) + " " + strings.TrimSpace(string(ver)), nil
}

func readFileSafe(path string) ([]byte, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	buf := make([]byte, 1000000)
	n, err := f.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
