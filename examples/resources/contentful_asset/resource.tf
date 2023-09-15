resource "contentful_asset" "example_asset" {
  asset_id = "test_asset"
  locale   = "en-US"
  space_id = "space-id"

  fields {
    title {
      locale  = "en-US"
      content = "asset title"
    }
    description {
      locale  = "en-US"
      content = "asset description"
    }
    file = {
      upload      = "https://images.ctfassets.net/fo9twyrwpveg/2VQx7vz73aMEYi20MMgCk0/66e502115b1f1f973a944b4bd2cc536f/IC-1H_Modern_Stack_Website.svg"
      fileName    = "example.jpeg"
      contentType = "image/jpeg"
    }
  }
  published = false
  archived  = false
}