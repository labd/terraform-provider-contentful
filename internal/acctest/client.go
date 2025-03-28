package acctest

import (
	"os"

	"github.com/labd/terraform-provider-contentful/internal/sdk"
	"github.com/labd/terraform-provider-contentful/internal/utils"
)

func GetClient() *sdk.ClientWithResponses {
	cmaToken := os.Getenv("CONTENTFUL_MANAGEMENT_TOKEN")

	client, err := utils.CreateClient("https://api.contentful.com", cmaToken)
	if err != nil {
		panic(err)
	}
	return client
}
