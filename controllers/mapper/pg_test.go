package mapper

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

	"github.com/k8gb-io/k8gb-light/controllers/utils"

	"github.com/stretchr/testify/assert"
)

func TestNewPrimaryGeoTag(t *testing.T) {
	var tests = []struct {
		name              string
		primaryGeoTag     string
		extClusterGeoTags []string
		clusterGeoTag     string
		expected          PrimaryGeotag
	}{
		{name: "US,EU from US EU ZA UK CZ", primaryGeoTag: "us, eu", extClusterGeoTags: []string{"us", "eu", "za", "uk", "cz"},
			clusterGeoTag: "us", expected: PrimaryGeotag{"us", "eu", "cz", "uk", "za"}},
		{name: "US,EU from EU US", primaryGeoTag: "us, eu", extClusterGeoTags: []string{"eu", "us"},
			clusterGeoTag: "us", expected: PrimaryGeotag{"us", "eu"}},
		{name: "US from US", primaryGeoTag: "us", extClusterGeoTags: []string{"us"},
			clusterGeoTag: "us", expected: PrimaryGeotag{"us"}},
		{name: "EMPTY of US, EU", primaryGeoTag: "", extClusterGeoTags: []string{"us", "eu"},
			clusterGeoTag: "us", expected: PrimaryGeotag{"eu", "us"}},
		{name: "ZA of US, EU", primaryGeoTag: "za", extClusterGeoTags: []string{"us", "eu"},
			clusterGeoTag: "us", expected: PrimaryGeotag{"za", "eu", "us"}},
		{name: "ZA of US, EU and DE", primaryGeoTag: "za", extClusterGeoTags: []string{"us", "eu"},
			clusterGeoTag: "de", expected: PrimaryGeotag{"za", "de", "eu", "us"}},
		{name: "DE,US of EU,US,ZA and DE", primaryGeoTag: "de, us", extClusterGeoTags: []string{"us", "eu", "za"},
			clusterGeoTag: "de", expected: PrimaryGeotag{"de", "us", "eu", "za"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// arrange
			// act
			rs := LoopState{Spec: Spec{PrimaryGeoTag: test.primaryGeoTag}}
			r := rs.GetFailoverOrderedGeotagList(test.clusterGeoTag, test.extClusterGeoTags)
			// assert
			b := utils.EqualItemsHasSameOrder(test.expected, r)
			assert.True(t, b)
		})
	}
}
