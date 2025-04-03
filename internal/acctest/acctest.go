package acctest

import (
	"net/http"
	"os"
	"testing"

	"github.com/labd/terraform-provider-contentful/internal/utils"
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

func TestHasNoContentTypes(t *testing.T) {
	client := GetClient()
	ctx := t.Context()

	spaceID := os.Getenv("CONTENTFUL_SPACE_ID")
	environment := os.Getenv("CONTENTFUL_ENVIRONMENT")
	if environment == "" {
		environment = "master"
	}

	resp, err := client.GetAllContentTypesWithResponse(ctx, spaceID, environment, nil)
	if err := utils.CheckClientResponse(resp, err, http.StatusOK); err != nil {
		t.Fatalf("error checking content types: %v", err)
	}

	if resp.JSON200.Total != 0 {
		t.Fatalf("expected no content types, but found %d", resp.JSON200.Total)
	}
}
