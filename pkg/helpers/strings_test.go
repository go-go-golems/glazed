package helpers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToAlphaString(t *testing.T) {
	// Test edge case where n is less than or equal to 0
	assert.Equal(t, "", ToAlphaString(0))
	assert.Equal(t, "", ToAlphaString(-1))

	// Test edge case where n is greater than 26
	assert.Equal(t, "AA", ToAlphaString(27))

	// Test edge case where n is exactly 26
	assert.Equal(t, "Z", ToAlphaString(26))

	// Test edge case where n is between 1 and 26
	assert.Equal(t, "A", ToAlphaString(1))
	assert.Equal(t, "J", ToAlphaString(10))

	// Test edge case where n is a large number
	assert.Equal(t, "CV", ToAlphaString(100))
}
