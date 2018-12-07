package collector

import (
	"testing"
)

func TestAWSCollectorCreation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
	// Currently only testing with default partition credentials
	for _, testCase := range []struct {
		partition string
		err       bool
	}{
		{partition: "default", err: false},
		{partition: "china", err: false},
		{partition: "non-existent", err: true},
	} {
		_, err := NewAWSCollector(testCase.partition)
		if have := (err != nil); testCase.err != have {
			t.Errorf("%s\tWant:%t\tHave:%t", testCase.partition, testCase.err, have)
		}
	}
}
