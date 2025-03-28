package contentful

import (
	"os"
)

var (
	// Environment variables
	spaceID  = os.Getenv("CONTENTFUL_SPACE_ID")
	CMAToken = os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	orgID    = os.Getenv("CONTENTFUL_ORGANIZATION_ID")

	// Terraform configuration values
	logBoolean = os.Getenv("TF_LOG")
)
