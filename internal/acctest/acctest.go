package acctest

import (
	"os"
	"testing"
)

func TestAccPreCheck(t *testing.T) {
	requiredEnvs := []string{
		"CONTENTFUL_MANAGEMENT_TOKEN",
		"CONTENTFUL_ORGANIZATION_ID",
		"CONTENTFUL_SPACE_ID",
	}
	for _, val := range requiredEnvs {
		if os.Getenv(val) == "" {
			t.Fatalf("%v must be set for acceptance tests", val)
		}
	}
}
