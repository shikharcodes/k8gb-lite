package controllers

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
	"fmt"
	"strings"

	"github.com/k8gb-io/k8gb-light/controllers/depresolver"
	"github.com/k8gb-io/k8gb-light/controllers/mapper"
	"github.com/k8gb-io/k8gb-light/controllers/providers/assistant"
	"github.com/k8gb-io/k8gb-light/controllers/providers/metrics"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	externaldns "sigs.k8s.io/external-dns/endpoint"
)

func (r *AnnoReconciler) getDNSEndpoint(rs *mapper.LoopState) (*externaldns.DNSEndpoint, error) {

	var gslbHosts []*externaldns.Endpoint
	var ttl = externaldns.TTL(rs.Spec.DNSTtlSeconds)

	localTargets, err := rs.GetExposedIPs()
	if err != nil {
		return nil, err
	}

	status := rs.GetStatus()
	for host, health := range status.ServiceHealth {
		var finalTargets = assistant.NewTargets()

		if !strings.Contains(host, r.Config.EdgeDNSZone) {
			return nil, fmt.Errorf("ingress host %s does not match delegated zone %s", host, r.Config.EdgeDNSZone)
		}

		isPrimary := false
		isHealthy := health == metrics.Healthy

		if isHealthy {
			finalTargets.Append(r.Config.ClusterGeoTag, localTargets)
			localTargetsHost := fmt.Sprintf("localtargets-%s", host)
			dnsRecord := &externaldns.Endpoint{
				DNSName:    localTargetsHost,
				RecordTTL:  ttl,
				RecordType: "A",
				Targets:    localTargets,
			}
			gslbHosts = append(gslbHosts, dnsRecord)
		}

		// Check if host is alive on external Gslb
		externalTargets := r.DNSProvider.GetExternalTargets(host)
		if len(externalTargets) > 0 {
			switch rs.Spec.Type {
			case depresolver.RoundRobinStrategy, depresolver.GeoStrategy:
				externalTargets.Sort()
				finalTargets.AppendTargets(externalTargets)
			case depresolver.FailoverStrategy:
				// If cluster is Primary and Healthy return only own targets
				// If cluster is Primary and Unhealthy return first Secondary Healthy cluster
				var topGeoTag string
				finalTargets.AppendTargets(externalTargets)
				primaryGeoTagList := rs.GetFailoverOrderedGeotagList(r.Config.ClusterGeoTag, r.Config.ExtClustersGeoTags)
				finalTargets, topGeoTag = finalTargets.FailoverProjection(primaryGeoTagList)
				isPrimary = topGeoTag == r.Config.ClusterGeoTag
				if isPrimary {
					if !isHealthy {
						r.Log.Info().
							Str("gslb", rs.NamespacedName.Name).
							Str("cluster", rs.Spec.PrimaryGeoTag).
							Strs("targets", finalTargets.GetIPs()).
							Str("workload", metrics.Unhealthy.String()).
							Msg("Executing failover strategy for primary cluster")
					}
				} else {
					r.Log.Info().
						Str("gslb", rs.NamespacedName.Name).
						Str("cluster", rs.Spec.PrimaryGeoTag).
						Strs("targets", finalTargets.GetIPs()).
						Str("workload", metrics.Healthy.String()).
						Msg("Executing failover strategy for secondary cluster")
				}
			}
		} else {
			r.Log.Info().
				Str("host", host).
				Msg("No external targets have been found for host")
		}

		r.updateRuntimeStatus(rs, isPrimary, health, finalTargets.GetIPs())
		r.Log.Info().
			Str("gslb", rs.NamespacedName.Name).
			Strs("targets", finalTargets.GetIPs()).
			Msg("Final target list")

		if len(finalTargets) > 0 {
			dnsRecord := &externaldns.Endpoint{
				DNSName:    host,
				RecordTTL:  ttl,
				RecordType: "A",
				Targets:    finalTargets.GetIPs(),
				Labels: externaldns.Labels{
					"strategy": rs.Spec.Type,
				},
			}
			for k, v := range r.getLabels(rs, finalTargets) {
				dnsRecord.Labels[k] = v
			}
			gslbHosts = append(gslbHosts, dnsRecord)
		}
	}
	dnsEndpointSpec := externaldns.DNSEndpointSpec{
		Endpoints: gslbHosts,
	}

	dnsEndpoint := &externaldns.DNSEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:        rs.NamespacedName.Name,
			Namespace:   rs.NamespacedName.Namespace,
			Annotations: map[string]string{"k8gb.absa.oss/dnstype": "local"},
			Labels:      map[string]string{"k8gb.absa.oss/dnstype": "local"},
		},
		Spec: dnsEndpointSpec,
	}

	err = controllerutil.SetControllerReference(rs.Ingress, dnsEndpoint, r.Scheme)
	if err != nil {
		return nil, err
	}
	return dnsEndpoint, err
}

// getLabels map of where key identifies region and weight, value identifies IP.
func (r *AnnoReconciler) getLabels(rs *mapper.LoopState, targets assistant.Targets) (labels map[string]string) {
	labels = make(map[string]string, 0)
	for k, v := range rs.Spec.Weights {
		t, found := targets[k]
		if !found {
			continue
		}
		for i, ip := range t.IPs {
			l := fmt.Sprintf("weight-%s-%v-%v", k, i, v)
			labels[l] = ip
		}
	}
	return labels
}

func (r *AnnoReconciler) updateRuntimeStatus(
	rs *mapper.LoopState,
	isPrimary bool,
	isHealthy metrics.HealthStatus,
	finalTargets []string,
) {
	switch rs.Spec.Type {
	case depresolver.RoundRobinStrategy:
		r.Metrics.UpdateRoundrobinStatus(rs.NamespacedName, isHealthy, finalTargets)
	case depresolver.GeoStrategy:
		r.Metrics.UpdateGeoIPStatus(rs.NamespacedName, isHealthy, finalTargets)
	case depresolver.FailoverStrategy:
		r.Metrics.UpdateFailoverStatus(rs.NamespacedName, isPrimary, isHealthy, finalTargets)
	}
}
