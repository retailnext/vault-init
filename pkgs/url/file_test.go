package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLocalFilePath(t *testing.T) {
	type filePathTestCases struct {
		description  string
		inputString  string
		expectedPath string
		expectErr    bool
	}
	for _, testCase := range []filePathTestCases{
		{
			description:  "valid absolute path",
			inputString:  "file:///absolute.txt",
			expectedPath: "/absolute.txt",
			expectErr:    false,
		},
		{
			description:  "valid absolute path 2",
			inputString:  "file:/path1/absolute.txt",
			expectedPath: "/path1/absolute.txt",
			expectErr:    false,
		},
		{
			description:  "invalid",
			inputString:  "file://absolute.txt",
			expectedPath: "",
			expectErr:    true,
		},
		{
			description:  "with host",
			inputString:  "file://localhost/absolute.txt",
			expectedPath: "/absolute.txt",
			expectErr:    false,
		},
	} {
		t.Run(testCase.description, func(t *testing.T) {
			filePath, err := GetLocalFilePath(testCase.inputString)
			if testCase.expectErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, testCase.expectedPath, filePath)
		})
	}
}
