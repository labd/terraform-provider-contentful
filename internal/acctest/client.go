package acctest

import (
	"github.com/flaconi/contentful-go"
	"os"
)

func GetClient() *contentful.Client {
	cmaToken := os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	organizationId := os.Getenv("CONTENTFUL_ORGANIZATION_ID")
	cma := contentful.NewCMA(cmaToken)
	cma.SetOrganization(organizationId)

	//cma.Debug = true
	return cma
}
