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

func Contains[T comparable](list []T, s T) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func Remove[T comparable](list []T, s T) []T {
	result := make([]T, 0)
	for _, v := range list {
		if v == s {
			continue
		}
		result = append(result, v)
	}
	return result
}

func EqualItems[T comparable](a, b []T) bool {
	// returns false when first slice is empty and second nil
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	x := make(map[T]bool, len(a))
	for _, ia := range a {
		x[ia] = false
	}
	for _, ib := range b {
		if _, found := x[ib]; found {
			x[ib] = true
		}
	}

	for _, v := range x {
		if !v {
			return false
		}
	}
	return true
}

// EqualItemsHasSameOrder two slices must have same values and order
func EqualItemsHasSameOrder[T comparable](a, b []T) bool {
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

// Merge multiple slices into one
func Merge[T any](x ...[]T) (y []T) {
	for _, v := range x {
		y = append(y, v...)
	}
	return y
}

// MergeWithSlice appends element y if doesn't exist in the x
func MergeWithSlice[T comparable](x []T, y ...T) (z []T) {
	m := AsMap(x)
	z = make([]T, len(x))
	copy(z, x)
	for _, v := range y {
		if !m[v] {
			z = append(z, v)
		}
	}
	return z
}

// MapHasOnlyKeys check that keys of map are identical to values in slice. If slice has different value than map
// or slice doesn't have item which exists in map (or vice versa), the program exits
func MapHasOnlyKeys[T comparable, U any](m map[T]U, x ...T) bool {
	if len(m) != len(x) {
		return false
	}
	mm := make(map[T]bool, len(m))
	for k := range m {
		mm[k] = false
	}

	for _, v := range x {
		if _, found := m[v]; !found {
			return false
		}
		mm[v] = true
	}

	for _, b := range mm {
		if !b {
			return false
		}
	}
	return true
}

// AsMap converts slice to map[T]int, where value is an index
func AsMap[T comparable](s []T) map[T]bool {
	m := make(map[T]bool, len(s))
	for _, v := range s {
		m[v] = true
	}
	return m
}
