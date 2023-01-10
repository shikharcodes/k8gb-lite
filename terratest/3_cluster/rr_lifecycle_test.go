package terratest

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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kuritka/annotation-operator/terratest"
	"github.com/kuritka/annotation-operator/terratest/utils"
)

func TestRoundRobinLifecycleOnThreeClusters(t *testing.T) {
	const ingressPath = "./resources/ingress_rr.yaml"
	const ingressEmptyPath = "./resources/ingress_empty.yaml"
	const digHits = 300
	const wgetHits = 150
	const expectedDigProbabilityDiff = 8
	const expectedWgetProbabilityDiff = 35
	instanceEU, err := utils.NewWorkflow(t, terratest.Environment.EUCluster, terratest.Environment.EUClusterPort).
		WithIngress(ingressPath).
		WithTestApp(terratest.Environment.EUCluster).
		WithBusybox().
		Start()
	require.NoError(t, err)
	defer instanceEU.Kill()

	instanceUS, err := utils.NewWorkflow(t, terratest.Environment.USCluster, terratest.Environment.USClusterPort).
		WithIngress(ingressPath).
		WithTestApp(terratest.Environment.USCluster).
		WithBusybox().
		Start()
	require.NoError(t, err)
	defer instanceUS.Kill()

	instanceZA, err := utils.NewWorkflow(t, terratest.Environment.ZACluster, terratest.Environment.ZAClusterPort).
		WithIngress(ingressPath).
		WithTestApp(terratest.Environment.ZACluster).
		WithBusybox().
		Start()
	require.NoError(t, err)
	defer instanceZA.Kill()

	allClusterIPs := utils.Merge(instanceEU.GetInfo().NodeIPs, instanceUS.GetInfo().NodeIPs, instanceZA.GetInfo().NodeIPs)

	t.Run("Wait until EU, US, ZA clusters are ready", func(t *testing.T) {
		// waiting until all localDNSEndpoints has all addresses
		err = instanceEU.Resources().WaitUntilDNSEndpointContainsTargets(instanceEU.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
		err = instanceUS.Resources().WaitUntilDNSEndpointContainsTargets(instanceUS.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
		err = instanceZA.Resources().WaitUntilDNSEndpointContainsTargets(instanceZA.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
	})

	t.Logf("All clusters are running 🚜💨! \n🇪🇺 %s;\n🇺🇲 %s;\n🇿🇦 %s",
		terratest.Environment.EUCluster,
		terratest.Environment.USCluster,
		terratest.Environment.ZACluster)

	t.Run("Digging one cluster, returned IPs of EU,US,ZA with the same probability", func(t *testing.T) {
		ips := instanceEU.Tools().DigNCoreDNS(digHits)
		p := ips.HasSimilarProbabilityOnPrecision(expectedDigProbabilityDiff)
		require.True(t, p, "Dig must return IPs with equal probability")
		require.True(t, utils.MapHasOnlyKeys(ips, allClusterIPs...))
	})

	t.Run("Wget application, EU,US,ZA clusters have similar probability", func(t *testing.T) {
		instanceHit := instanceEU.Tools().WgetNTestApp(wgetHits)
		p := instanceHit.HasSimilarProbabilityOnPrecision(expectedWgetProbabilityDiff)
		require.True(t, p, "Instance Hit must return clusters with similar probability")
		require.True(t, utils.MapHasOnlyKeys(instanceHit, terratest.Environment.EUCluster, terratest.Environment.USCluster,
			terratest.Environment.ZACluster))
	})

	allClusterIPs = utils.Merge(instanceEU.GetInfo().NodeIPs, instanceUS.GetInfo().NodeIPs)
	t.Run("Killing ZA App, ZA ingress status is Unhealthy", func(t *testing.T) {
		instanceZA.App().StopTestApp()
		// waiting until all localDNSEndpoints has all addresses
		err = instanceEU.Resources().WaitUntilDNSEndpointContainsTargets(instanceEU.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
		err = instanceUS.Resources().WaitUntilDNSEndpointContainsTargets(instanceUS.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
		err = instanceZA.Resources().WaitUntilDNSEndpointContainsTargets(instanceZA.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
	})

	t.Run("Digging one cluster, returned IPs of EU, US with the same probability", func(t *testing.T) {
		ips := instanceUS.Tools().DigNCoreDNS(digHits)
		p := ips.HasSimilarProbabilityOnPrecision(expectedDigProbabilityDiff)
		require.True(t, p, "Dig must return IPs with equal probability")
		require.True(t, utils.MapHasOnlyKeys(ips, allClusterIPs...))
	})

	t.Run("Wget application, EU,US clusters have similar probability", func(t *testing.T) {
		instanceHit := instanceEU.Tools().WgetNTestApp(wgetHits)
		p := instanceHit.HasSimilarProbabilityOnPrecision(expectedWgetProbabilityDiff)
		require.True(t, p, "Instance Hit must return clusters with similar probability")
		require.True(t, utils.MapHasOnlyKeys(instanceHit, terratest.Environment.EUCluster, terratest.Environment.USCluster))
	})

	allClusterIPs = utils.Merge(instanceUS.GetInfo().NodeIPs)
	t.Run("Killing EU Namespace, EU ingress and App doesnt exists", func(t *testing.T) {
		instanceEU.Kill()
		err = instanceUS.Resources().WaitUntilDNSEndpointContainsTargets(instanceUS.GetInfo().Host, allClusterIPs)
		require.NoError(t, err, "WARNING: If you running test locally, ensure the App IS NOT running in forgotten namespaces")
		err = instanceZA.Resources().WaitUntilDNSEndpointContainsTargets(instanceZA.GetInfo().Host, allClusterIPs)
		require.NoError(t, err, "WARNING: If you running test locally, ensure the App IS NOT running in forgotten namespaces")
	})

	t.Run("Digging one cluster, returns IPs of US cluster", func(t *testing.T) {
		ips := instanceUS.Tools().DigNCoreDNS(20)
		require.True(t, utils.MapHasOnlyKeys(ips, allClusterIPs...))
	})

	t.Run("Wget application US clusters have similar probability", func(t *testing.T) {
		instanceHit := instanceUS.Tools().WgetNTestApp(10)
		require.True(t, utils.MapHasOnlyKeys(instanceHit, terratest.Environment.USCluster))
	})

	allClusterIPs = []string{}
	t.Run("ReApply US ingress, remove K8gb annotation", func(t *testing.T) {
		instanceUS.ReapplyIngress(ingressEmptyPath)

		// US dns endpoint not found now
		err = instanceUS.Resources().WaitUntilDNSEndpointNotFound()
		require.NoError(t, err)
	})

	t.Logf("All is broken!🧨 \n🇪🇺 %s is removed;\n🇺🇲 %s has no k8gb annotation;\n🇿🇦 %s has stopped app",
		terratest.Environment.EUCluster,
		terratest.Environment.USCluster,
		terratest.Environment.ZACluster)

	t.Logf("Spinnig all of them back! 🎩🍀")
	allClusterIPs = utils.Merge(instanceZA.GetInfo().NodeIPs)
	t.Run("Starting ZA App, ZA ingress status is Healthy", func(t *testing.T) {
		instanceZA.App().StartTestApp()
		// waiting until all localDNSEndpoints has all addresses
		err = instanceUS.Resources().WaitUntilDNSEndpointNotFound()
		require.NoError(t, err)
		err = instanceZA.Resources().WaitUntilDNSEndpointContainsTargets(instanceZA.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
	})

	t.Run("Digging one cluster, returns IPs of ZA cluster", func(t *testing.T) {
		ips := instanceZA.Tools().DigNCoreDNS(20)
		require.True(t, utils.MapHasOnlyKeys(ips, allClusterIPs...))
	})

	t.Run("Wget application ZA clusters ", func(t *testing.T) {
		instanceHit := instanceZA.Tools().WgetNTestApp(10)
		require.True(t, utils.MapHasOnlyKeys(instanceHit, terratest.Environment.ZACluster))
	})

	t.Run("Reapply US ingress(add K8gb annotation) and recreate EU namespace", func(t *testing.T) {
		instanceUS.ReapplyIngress(ingressPath)
		instanceEU, err = utils.NewWorkflow(t, terratest.Environment.EUCluster, terratest.Environment.EUClusterPort).
			WithIngress(ingressPath).
			WithTestApp(terratest.Environment.EUCluster).
			WithBusybox().
			Start()
		require.NoError(t, err)

		allClusterIPs = utils.Merge(instanceZA.GetInfo().NodeIPs, instanceUS.GetInfo().NodeIPs, instanceEU.GetInfo().NodeIPs)
		// US dns endpoint not found now
		err = instanceZA.Resources().WaitUntilDNSEndpointContainsTargets(instanceZA.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
		err = instanceUS.Resources().WaitUntilDNSEndpointContainsTargets(instanceUS.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
		err = instanceEU.Resources().WaitUntilDNSEndpointContainsTargets(instanceUS.GetInfo().Host, allClusterIPs)
		require.NoError(t, err)
	})

	t.Run("Digging one cluster, returned IPs of EU,US,ZA with the same probability", func(t *testing.T) {
		ips := instanceZA.Tools().DigNCoreDNS(digHits)
		p := ips.HasSimilarProbabilityOnPrecision(expectedDigProbabilityDiff)
		require.True(t, p, "Dig must return IPs with equal probability")
		require.True(t, utils.MapHasOnlyKeys(ips, allClusterIPs...))
	})

	t.Run("Wget application, EU,US,ZA clusters have similar probability", func(t *testing.T) {
		instanceHit := instanceUS.Tools().WgetNTestApp(wgetHits)
		p := instanceHit.HasSimilarProbabilityOnPrecision(expectedWgetProbabilityDiff)
		require.True(t, p, "Instance Hit must return clusters with similar probability")
		require.True(t, utils.MapHasOnlyKeys(instanceHit, terratest.Environment.EUCluster, terratest.Environment.USCluster,
			terratest.Environment.ZACluster))
	})

}
