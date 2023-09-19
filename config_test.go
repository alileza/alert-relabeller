package main

import (
	"reflect"
	"testing"

	"github.com/prometheus/common/model"
)

func TestSanitized(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		// Test case with alphanumeric characters only
		{"Hello123", "hello123"},

		// Test case with spaces
		{"Hello World", "helloworld"},

		// Test case with special characters
		{"This-is_a@Test!", "thisisatest"},

		// Test case with quotes
		{"'Quoted Text'", "quotedtext"},

		// Test case with mixed characters
		{"123 456!ABC", "123456abc"},

		{"1`2\"3' 456!ABC", "123456abc"},

		// Test case with empty input
		{"", ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			result := sanitize(testCase.input)
			if result != testCase.expected {
				t.Errorf("Input: %s, Expected: %s, Got: %s", testCase.input, testCase.expected, result)
			}
		})
	}
}

func TestParseCondition(t *testing.T) {
	testCases := []struct {
		input    string
		expected []model.LabelValue
		isError  bool
	}{
		// Valid input case
		{
			input:    "key==value",
			expected: []model.LabelValue{"key", "value"},
			isError:  false,
		},
		// Invalid input case
		{
			input:    "invalid_input",
			expected: []model.LabelValue{},
			isError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := parseCondition(tc.input)

			if tc.isError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if !reflect.DeepEqual(result, tc.expected) {
					t.Errorf("Expected %v, but got %v", tc.expected, result)
				}
			}
		})
	}
}
