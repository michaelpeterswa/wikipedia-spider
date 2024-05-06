package rules_test

import (
	"testing"

	"github.com/michaelpeterswa/wikipedia-spider/internal/rules"
	"github.com/stretchr/testify/assert"
)

func TestIsValidLink(t *testing.T) {
	tests := []struct {
		name     string
		link     string
		expected bool
	}{
		{
			name:     "basic test fail",
			link:     "/wiki/Category:Test",
			expected: false,
		},
		{
			name:     "basic test fail 2",
			link:     "/wiki/Special:Test",
			expected: false,
		},
		{
			name:     "basic test fail 3",
			link:     "/wiki/Help:Test",
			expected: false,
		},
		{
			name:     "basic test fail 4",
			link:     "/wiki/Wikipedia:Test",
			expected: false,
		},
		{
			name:     "basic test fail 5",
			link:     "/wiki/Template:Test",
			expected: false,
		},
		{
			name:     "basic test fail 6",
			link:     "/wiki/Template_talk:Test",
			expected: false,
		},
		{
			name:     "basic test fail 7",
			link:     "/wiki/File:Test.png",
			expected: false,
		},
		{
			name:     "basic test fail 8",
			link:     "#cite_note-37",
			expected: false,
		},
		{
			name:     "basic test pass",
			link:     "/wiki/Kingman_Reef",
			expected: true,
		},
		{
			name:     "basic test pass",
			link:     "/wiki/San_Diego",
			expected: true,
		},
		{
			name:     "basic test pass",
			link:     "/wiki/Kingman_Reef",
			expected: true,
		},
		{
			name:     "basic test pass",
			link:     "/wiki/Idaho",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, rules.IsValidLink(tc.link))
		})
	}
}
