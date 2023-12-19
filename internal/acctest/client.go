package acctest

import (
	"github.com/flaconi/contentful-go"
	client2 "github.com/flaconi/contentful-go/pkgs/client"
	"github.com/flaconi/contentful-go/service/common"
	"os"
)

func GetClient() *contentful.Client {
	cmaToken := os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")
	organizationId := os.Getenv("CONTENTFUL_ORGANIZATION_ID")
	cma := contentful.NewCMA(cmaToken)
	cma.SetOrganization(organizationId)

	return cma
}

func GetCMA() common.SpaceIdClientBuilder {
	client, err := contentful.NewCMAV2(client2.ClientConfig{
		URL:       "https://api.contentful.com",
		Debug:     false,
		UserAgent: "terraform-provider-contentful-test",
		Token:     os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN"),
	})

	if err != nil {
		panic(err)
	}

	return client
}
