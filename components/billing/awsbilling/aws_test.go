package awsbilling

import (
	"testing"
)

func TestGet(t *testing.T) {
	ab := AWS{
		AccessKeyID:     "",
		SecretAccessKey: "",
		Region:          "us-east-1",
		Currency:        "USD",
	}

	ab.Get()
}
