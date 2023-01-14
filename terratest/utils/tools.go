package utils

/*
Copyright 2022 The k8gb Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/shell"
)

type HitCount map[string]int

type Tools struct {
	i *Instance
}

// DigCoreDNSHost digs CoreDNS for cluster instance
func (t *Tools) DigCoreDNSHost(host string) []string {
	port := fmt.Sprintf("-p%d", t.i.w.port)
	dnsServer := fmt.Sprintf("@%s", "localhost")
	digApp := shell.Command{
		Command: "dig",
		Args:    []string{port, dnsServer, host, "+short", "+tcp", "-4"},
	}
	digAppOut := shell.RunCommandAndGetOutput(t.i.w.t, digApp)
	if digAppOut == "" {
		return []string{}
	}
	return strings.Split(digAppOut, "\n")
}

// DigCoreDNS digs CoreDNS for cluster instance
func (t *Tools) DigCoreDNS() []string {
	return t.DigCoreDNSHost(t.i.GetInfo().Host)
}

// DigNCoreDNS digs CoreDNS for cluster instance
func (t *Tools) DigNCoreDNS(n int) HitCount {
	m := make(HitCount, 0)
	for i := 0; i < n; i++ {
		ips := t.DigCoreDNS()
		if len(ips) > 0 {
			m[ips[0]]++
		}
		//for _, ip := range ips {
		//	m[ip]++
		//}
	}
	return m
}

func (t *Tools) WgetNTestApp(n int) HitCount {
	m := make(HitCount, 0)
	for i := 0; i < n; i++ {
		m[t.WgetTestApp()]++
	}
	delete(m, "err")
	return m
}

func (t *Tools) WgetTestApp() string {
	require.True(t.i.w.t, t.i.w.busybox.isRunning, "Busybox needs to be running. Use WithBusybox() function when init cluster")
	// require.True(t.i.w.t, t.i.w.testApp.isRunning, "TestApp needs to be running. Use WithTestApp(...) function when init cluster")
	data := struct {
		Message string `json:"message"`
	}{}
	host := t.i.GetInfo().Host
	cmd := shell.Command{
		Command: "kubectl",
		Args:    []string{"--context", t.i.w.k8sOptions.ContextName, "-n", t.i.w.namespace, "exec", "-i", "busybox", "--", "wget", "-qO", "-", host},
		Env:     t.i.w.k8sOptions.Env,
	}
	out, err := shell.RunCommandAndGetOutputE(t.i.w.t, cmd)
	if out == "Error from server: error dialing backend: EOF" {
		appStatus := t.i.getAppStatus()
		nodesIPs := t.i.getNodesIPs()
		t.i.w.t.Logf("NodeIPs %v; AppStatus: %v", nodesIPs, appStatus)
		return "err"
	}
	require.NoError(t.i.w.t, err)
	err = json.Unmarshal([]byte(out), &data)
	require.NoError(t.i.w.t, err)
	return data.Message
}

// HasSimilarProbabilityOnPrecision : ip addresses will appear in the map with a certain probability that is away
// from the average by the deviationPercentage value. For example, we have 400 requests and 4 IP addresses.
// If deviationPercentage =5%, then one address may have 103, the second 97, the third 105 and the fourth 95. Returns true.
// If deviationPercentage =5%, then one address may have 106 hits, the function returns false.
func (f HitCount) HasSimilarProbabilityOnPrecision(deviationPercentage int) bool {
	var r float64
	for _, v := range f {
		r += float64(v)
	}
	r = r / float64(len(f))
	da := r * float64(100-deviationPercentage) / 100
	db := r * float64(100+deviationPercentage) / 100
	for _, v := range f {
		if float64(v) < da {
			return false
		}
		if float64(v) > db {
			return false
		}
	}
	return true
}

func (f HitCount) Append(hitCount HitCount) (h HitCount) {
	h = make(map[string]int)
	for k, v := range f {
		h[k] = v
	}
	for k, v := range hitCount {
		h[k] += v
	}
	return h
}

func (f HitCount) HasExpectedProbabilityWithPrecision(i *Instance, expectedProbabilityPercentage, deviationPercentage int) bool {
	ips := i.GetInfo().NodeIPs
	var total, hits int
	for _, v := range ips {
		hits += f[v]
	}
	for _, v := range f {
		total += v
	}
	ratio := hits * 100 / total

	dl := expectedProbabilityPercentage - deviationPercentage
	dh := expectedProbabilityPercentage + deviationPercentage

	return ratio < dh && ratio > dl
}
