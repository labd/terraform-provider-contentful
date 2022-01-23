package contentful

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	contentful "github.com/labd/contentful-go"
)

func TestAccContentfulLocales_Basic(t *testing.T) {
	var locale contentful.Locale

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccContentfulLocaleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccContentfulLocaleConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulLocaleExists("contentful_locale.mylocale", &locale),
					testAccCheckContentfulLocaleAttributes(&locale, map[string]interface{}{
						"space_id":      spaceID,
						"name":          "locale-name",
						"code":          "de",
						"fallback_code": "en-US",
						"optional":      false,
						"cda":           false,
						"cma":           true,
					}),
				),
			},
			{
				Config: testAccContentfulLocaleUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContentfulLocaleExists("contentful_locale.mylocale", &locale),
					testAccCheckContentfulLocaleAttributes(&locale, map[string]interface{}{
						"space_id":      spaceID,
						"name":          "locale-name-updated",
						"code":          "es",
						"fallback_code": "en-US",
						"optional":      true,
						"cda":           true,
						"cma":           false,
					}),
				),
			},
		},
	})
}

func testAccCheckContentfulLocaleExists(n string, locale *contentful.Locale) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not Found: %s", n)
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		localeID := rs.Primary.ID
		if localeID == "" {
			return fmt.Errorf("no locale ID is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		contentfulLocale, err := client.Locales.Get(spaceID, localeID)
		if err != nil {
			return err
		}

		*locale = *contentfulLocale

		return nil
	}
}

func testAccCheckContentfulLocaleAttributes(locale *contentful.Locale, attrs map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := attrs["name"].(string)
		if locale.Name != name {
			return fmt.Errorf("locale name does not match: %s, %s", locale.Name, name)
		}

		code := attrs["code"].(string)
		if locale.Code != code {
			return fmt.Errorf("locale code does not match: %s, %s", locale.Code, code)
		}

		fallbackCode := attrs["fallback_code"].(string)
		if locale.FallbackCode != fallbackCode {
			return fmt.Errorf("locale fallback code does not match: %s, %s", locale.FallbackCode, fallbackCode)
		}

		isOptional := attrs["optional"].(bool)
		if locale.Optional != isOptional {
			return fmt.Errorf("locale options value does not match: %t, %t", locale.Optional, isOptional)
		}

		isCDA := attrs["cda"].(bool)
		if locale.CDA != isCDA {
			return fmt.Errorf("locale cda does not match: %t, %t", locale.CDA, isCDA)
		}

		isCMA := attrs["cma"].(bool)
		if locale.CMA != isCMA {
			return fmt.Errorf("locale cma does not match: %t, %t", locale.CMA, isCMA)
		}

		return nil
	}
}

func testAccContentfulLocaleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "contentful_locale" {
			continue
		}

		spaceID := rs.Primary.Attributes["space_id"]
		if spaceID == "" {
			return fmt.Errorf("no space_id is set")
		}

		localeID := rs.Primary.ID
		if localeID == "" {
			return fmt.Errorf("no locale ID is set")
		}

		client := testAccProvider.Meta().(*contentful.Client)

		_, err := client.Locales.Get(spaceID, localeID)

		// This is caused by a bug in the client library that tries to cast
		// error.details to a map but it is a string in this case
		if _, ok := err.(*json.UnmarshalTypeError); ok {
			return nil
		}

		return fmt.Errorf("locale still exists with id: %s", localeID)
	}

	return nil
}

var testAccContentfulLocaleConfig = `
resource "contentful_locale" "mylocale" {
  space_id = "` + spaceID + `"

  name = "locale-name"
  code = "de"
  fallback_code = "en-US"
  optional = false
  cda = false
  cma = true
}
`

var testAccContentfulLocaleUpdateConfig = `
resource "contentful_locale" "mylocale" {
  space_id = "` + spaceID + `"

  name = "locale-name-updated"
  code = "es"
  fallback_code = "en-US"
  optional = true
  cda = true
  cma = false
}
`
