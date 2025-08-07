# Example demonstrating the new intuitive default_value syntax

resource "contentful_contenttype" "demo_content_type" {
  space_id      = "demo-space"
  environment   = "master"
  id            = "demo_type"
  name          = "Demo Content Type"
  description   = "Demonstrates the new default_value syntax"
  
  fields = [
    {
      id   = "simpleString"
      name = "Simple String"
      type = "Symbol"
      # Simple string default - will use en-US locale
      default_value = "Hello World"
      required = false
    },
    {
      id   = "simpleBool"
      name = "Simple Boolean"
      type = "Boolean"
      # Simple boolean default - will use en-US locale
      default_value = true
      required = false
    },
    {
      id   = "themeColor"
      name = "Theme Color"
      type = "Symbol"
      validations = [{
        in = ["green", "pink", "turquoise", "yellow", "purple"]
      }]
      # Simple string with validation
      default_value = "green"
      required = false
    },
    {
      id   = "localizedGreeting"
      name = "Localized Greeting"
      type = "Symbol"
      # Multiple locales using map syntax
      default_value = {
        "en-US" = "Hello"
        "de-DE" = "Hallo" 
        "es-ES" = "Hola"
        "fr-FR" = "Bonjour"
      }
      required = false
    }
  ]
}
