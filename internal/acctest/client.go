package acctest

import (
	"os"

	"github.com/labd/contentful-go"
	client2 "github.com/labd/contentful-go/pkgs/client"
	"github.com/labd/contentful-go/pkgs/util"
	"github.com/labd/contentful-go/service/cma"
)

func GetClient() *contentful.Client {
	cmaToken := os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	organizationId := os.Getenv("CONTENTFUL_ORGANIZATION_ID")
	cma := contentful.NewCMA(cmaToken)
	cma.SetOrganization(organizationId)

	return cma
}

func GetCMA() cma.SpaceIdClientBuilder {
	client, err := contentful.NewCMAV2(client2.ClientConfig{
		Debug:     false,
		UserAgent: util.ToPointer("terraform-provider-contentful-test"),
		Token:     os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN"),
	})

	if err != nil {
		panic(err)
	}

	return client
}
