package util

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

const (
	getDefaultNetworkCMD = "networksetup -listnetworkserviceorder | grep" +
		" `(route -n get default | grep 'interface' || route -n get -inet6 default | grep 'interface') | cut -d ':' -f2`" +
		" -B 1 | head -n 1 | cut -d ' ' -f 2-"

	darwinOS                     = "darwin"
	permissionErrorHelpTextMacOS = "By default SpoofDPI tries to set itself up as a system-wide proxy server.\n" +
		"Doing so may require root access on machines with\n" +
		"'Settings > Privacy & Security > Advanced > Require" +
		" an administrator password to access system-wide settings' enabled.\n" +
		"If you do not want SpoofDPI to act as a system-wide proxy, provide" +
		" -system-proxy=false."
)

// SetOsProxy sets the system proxy settings on macOS.
func SetOsProxy(port uint16) error {
	if runtime.GOOS != darwinOS {
		return nil
	}

	network, err := getDefaultNetwork()
	if err != nil {
		return err
	}

	return setProxy(getProxyTypes(), network, "127.0.0.1", port)
}

// UnsetOsProxy unsets the system proxy settings on macOS.
func UnsetOsProxy() error {
	if runtime.GOOS != darwinOS {
		return nil
	}

	network, err := getDefaultNetwork()
	if err != nil {
		return err
	}

	return unsetProxy(getProxyTypes(), network)
}

// getDefaultNetwork retrieves the default network interface on macOS.
func getDefaultNetwork() (string, error) {
	output, err := exec.Command("sh", "-c", getDefaultNetworkCMD).Output()
	if err != nil {
		return "", err
	}
	if len(output) == 0 {
		return "", errors.New("no available networks")
	}
	return strings.TrimSpace(string(output)), nil
}

// getProxyTypes returns the types of proxies that can be set on macOS.
func getProxyTypes() []string {
	return []string{"webproxy", "securewebproxy"}
}

// setProxy sets the proxy settings for the specified network interface.
func setProxy(proxyTypes []string, network, domain string, port uint16) error {
	args := []string{"", network, domain, strconv.FormatUint(uint64(port), 10)}

	for _, proxyType := range proxyTypes {
		args[0] = "-set" + proxyType
		err := networkSetup(args)
		if err != nil {
			return fmt.Errorf("setting %s: %w", proxyType, err)
		}
	}
	return nil
}

// unsetProxy disables the proxy settings for the specified network interface.
func unsetProxy(proxyTypes []string, network string) error {
	args := []string{"", network, "off"}

	for _, proxyType := range proxyTypes {
		args[0] = "-set" + proxyType + "state"
		err := networkSetup(args)
		if err != nil {
			return fmt.Errorf("unsetting %s: %w", proxyType, err)
		}
	}
	return nil
}

// networkSetup executes the networksetup command with the provided arguments.
func networkSetup(args []string) error {
	cmd := exec.Command("networksetup", args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}

	msg := string(out)
	if isMacOSPermissionError(err) {
		msg += permissionErrorHelpTextMacOS
	}
	return fmt.Errorf("%s: %s", cmd.String(), msg)
}

// isMacOSPermissionError checks if the error is a macOS-specific permission error.
func isMacOSPermissionError(err error) bool {
	if runtime.GOOS != darwinOS {
		return false
	}

	var exitErr *exec.ExitError
	return errors.As(err, &exitErr) && exitErr.ExitCode() == 14
}
