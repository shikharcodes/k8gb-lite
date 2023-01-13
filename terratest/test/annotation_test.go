package test

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
	"time"

	"github.com/kuritka/annotation-operator/terratest"
	"github.com/kuritka/annotation-operator/terratest/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnotation(t *testing.T) {
	const ingressPath1 = "./resources/ingress_annotation1.yaml"
	const ingressPath2 = "./resources/ingress_annotation2.yaml"
	const ingressPath3 = "./resources/ingress_annotation3.yaml"
	const ingressPath4 = "./resources/ingress_annotation4.yaml"
	instanceEU, err := utils.NewWorkflow(t, terratest.Environment.EUCluster, terratest.Environment.EUClusterPort).
		WithIngress(ingressPath1).
		WithTestApp("eu").
		Start()
	assert.NoError(t, err)
	defer instanceEU.Kill()

	host := instanceEU.GetInfo().Host
	ips := instanceEU.GetInfo().NodeIPs

	t.Run("k8gb.io/strategy is not found", func(t *testing.T) {
		err = instanceEU.Resources().WaitUntilDNSEndpointNotFound()
		require.NoError(t, err)
	})

	t.Run("k8gb.io/strategy is roundrobin, patching annotations", func(t *testing.T) {
		instanceEU.Resources().IngressPatchAnnotation("k8gb.io/terratest-patch", "#FC4")
		require.Equal(t, "#FC4", instanceEU.Resources().Ingress().Annotations["k8gb.io/terratest-patch"])
	})

	t.Run("k8gb.io/strategy is failover", func(t *testing.T) {
		instanceEU.ReapplyIngress(ingressPath2)
		err = instanceEU.Resources().WaitUntilDNSEndpointContainsTargets(host, ips)
		require.NoError(t, err)
		ep := instanceEU.Resources().GetLocalDNSEndpoint().GetEndpointByName(host)
		require.Equal(t, 30, ep.RecordTTL)
		require.Equal(t, 1, len(ep.Labels))
		require.Equal(t, "failover", ep.Labels["strategy"])
	})

	t.Run("k8gb.io/strategy is roundrobin", func(t *testing.T) {
		instanceEU.ReapplyIngress(ingressPath3)
		time.Sleep(5 * time.Second)
		ep := instanceEU.Resources().GetLocalDNSEndpoint().GetEndpointByName(host)
		require.Equal(t, 222, ep.RecordTTL)
		require.Equal(t, 1, len(ep.Labels))
		require.Equal(t, "roundRobin", ep.Labels["strategy"])
		require.Equal(t, "annotation-test3", instanceEU.Resources().Ingress().Annotations["xxx"])
		require.Equal(t, "eu", instanceEU.Resources().Ingress().Annotations["k8gb.io/primary-geotag"])

	})

	t.Run("k8gb.io/strategy is missing", func(t *testing.T) {
		instanceEU.ReapplyIngress(ingressPath4)
		err = instanceEU.Resources().WaitUntilDNSEndpointNotFound()
		require.NoError(t, err)
		require.Equal(t, "annotation-test4", instanceEU.Resources().Ingress().Annotations["xxx"])
		_, found := instanceEU.Resources().Ingress().Annotations["k8gb.io/primary-geotag"]
		require.False(t, found)
	})

}
