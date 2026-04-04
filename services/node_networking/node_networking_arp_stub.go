//go:build !linux

package node_networking

import (
	"fmt"
	"runtime"
	"time"
)

func (n *GoInfraRESTApiPinger) ArpPing(remoteHost string) PingResult {
	return PingResult{
		TargetHostName: remoteHost,
		Success:        false,
		Error:          fmt.Errorf("ARP ping is only supported on Linux, current platform is %s", runtime.GOOS),
		AverageLatency: time.Duration(0),
	}
}
