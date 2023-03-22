package list

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReverse(t *testing.T) {
	// Test edge case where s is empty
	s := []int{}
	Reverse(s)
	assert.Equal(t, []int{}, s)

	// Test edge case where s has a single element
	s = []int{1}
	Reverse(s)
	assert.Equal(t, []int{1}, s)

	// Test edge case where s has multiple elements
	s = []int{1, 2, 3, 4, 5}
	Reverse(s)
	assert.Equal(t, []int{5, 4, 3, 2, 1}, s)

	// with strings
	s2 := []string{"a", "b", "c", "d", "e"}
	Reverse(s2)
	assert.Equal(t, []string{"e", "d", "c", "b", "a"}, s2)
}
