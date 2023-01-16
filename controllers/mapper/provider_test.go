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
	"fmt"
	"reflect"
	"testing"

	"cloud.example.com/annotation-operator/controllers/utils"

	"cloud.example.com/annotation-operator/controllers/providers/metrics"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"cloud.example.com/annotation-operator/controllers/depresolver"

	"github.com/golang/mock/gomock"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetIngress(t *testing.T) {
	const (
		ns   = "test-namespace"
		name = "test-name"
	)
	var tests = []struct {
		name            string
		nn              types.NamespacedName
		expectedIngress *netv1.Ingress
		expectedResult  Result
		err             error
		strategy        string
	}{
		{name: "Client Error", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{}, expectedResult: ResultError, err: fmt.Errorf("reading resource error")},
		{name: "NotFound Selector", nn: types.NamespacedName{Name: name},
			expectedIngress: &netv1.Ingress{}, expectedResult: ResultNotFound, err: nil},
		{name: "Ingress Without Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{
				Name: name}}, expectedResult: ResultExistsButNotAnnotationFound, err: nil},
		{name: "Ingress With Unrelated Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{"x.y.com": "10"},
				Name: name}}, expectedResult: ResultExistsButNotAnnotationFound, err: nil},
		{name: "Ingress With RR Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.RoundRobinStrategy},
				Name: name}}, expectedResult: ResultExists, err: nil},
		{name: "Ingress With RR Annotation With Redundant PrimaryGeoTag", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.RoundRobinStrategy,
				AnnotationPrimaryGeoTag: "eu"}, Name: name}}, expectedResult: ResultExists, err: nil},
		{name: "Ingress With FO Invalid Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.FailoverStrategy},
				Name: name}}, expectedResult: ResultError, err: nil},
		{name: "Ingress With FO Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.FailoverStrategy,
				AnnotationPrimaryGeoTag: "eu"}, Name: name}}, expectedResult: ResultExists, err: nil},
		{name: "Ingress With WRR Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.RoundRobinStrategy,
				AnnotationWeightJSON: "eu:5,us:10"}, Name: name}}, expectedResult: ResultExists, err: nil},
		{name: "Ingress With WRR Invalid Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.RoundRobinStrategy,
				AnnotationWeightJSON: "eu:a,us:10"}, Name: name}}, expectedResult: ResultError, err: nil},
		{name: "Ingress With GeoIP", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: depresolver.GeoStrategy},
				Name: name}}, expectedResult: ResultExists, err: nil},
		{name: "Ingress With NONEXISTING Annotation", nn: types.NamespacedName{Name: name, Namespace: ns},
			expectedIngress: &netv1.Ingress{ObjectMeta: metav1.ObjectMeta{Annotations: map[string]string{AnnotationStrategy: "NON-EXISTING"},
				Name: name}}, expectedResult: ResultError, err: nil},
	}

	for _, test := range tests {

		// arrange
		t.Run(test.name, func(t *testing.T) {
			m := M(t)
			m.Client.(*MockClient).EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
				func(arg0, arg1 interface{}, ing *netv1.Ingress, args ...interface{}) error {
					if test.nn.Name == "" || test.nn.Namespace == "" {
						return errors.NewNotFound(schema.GroupResource{}, name)
					}
					ing.ObjectMeta = test.expectedIngress.ObjectMeta
					return test.err
				})

			// act
			rs, result, err := NewCommonProvider(m.Client, &depresolver.Config{}).Get(test.nn)

			// assert
			assert.Equal(t, test.expectedResult, result)
			assert.Equal(t, result == ResultError, err != nil)
			if result == ResultError {
				assert.Nil(t, rs)
			} else {
				assert.True(t, reflect.DeepEqual(test.expectedIngress.ObjectMeta, rs.Ingress.ObjectMeta))
				assert.IsType(t, rs.Mapper, &IngressMapper{})
				assert.True(t, reflect.DeepEqual(rs.Status, Status{ServiceHealth: map[string]metrics.HealthStatus{},
					HealthyRecords: map[string][]string{}, GeoTag: "", Hosts: ""}))
			}
		})
	}
}

func M(t *testing.T) struct {
	Client client.Client
	Dig    utils.Digger
} {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	return struct {
		Client client.Client
		Dig    utils.Digger
	}{
		Client: NewMockClient(ctrl),
		Dig:    NewMockDigger(ctrl),
	}
}
