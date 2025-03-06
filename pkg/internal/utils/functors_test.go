// functors_test.go file
package utils_test

import (
	"reflect"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

func TestMap(t *testing.T) {
	elems := []int{1, 2, 3, 4}
	doubledElems := utils.Map(elems, func(i int) int {
		return i * 2
	})

	expected := []int{2, 4, 6, 8}
	if !reflect.DeepEqual(doubledElems, expected) {
		t.Errorf("Expected %v, got %v", expected, doubledElems)
	}
}

func TestFilter(t *testing.T) {
	elems := []int{1, 2, 3, 4, 5, 6}
	filteredElems := utils.Filter(elems, func(i int) bool {
		return i%2 == 0 // Keep only even numbers
	})

	expected := []int{2, 4, 6}
	if !reflect.DeepEqual(filteredElems, expected) {
		t.Errorf("Expected %v, got %v", expected, filteredElems)
	}
}
