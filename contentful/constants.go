package contentful

import (
	"os"
	"time"
)

var (
	// Environment variables
	spaceID  = os.Getenv("SPACE_ID")
	CMAToken = os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	orgID    = os.Getenv("CONTENTFUL_ORGANIZATION_ID")

	// Terraform configuration values
	logBoolean = os.Getenv("TF_LOG")
)

var consistencyBuffer time.Duration = 10 * time.Second
