resource "contentful_taxonomy_concept" "example_concept" {
  space_id         = "your-space-id"
  environment      = "master"
  concept_scheme_id = "your-concept-scheme-id"
  
  pref_label = {
    en = "Technology"
    de = "Technologie"
  }
  
  alt_label = {
    en = ["Tech", "Technical"]
    de = ["Technik"]
  }
  
  definition = {
    en = "The application of scientific knowledge for practical purposes"
    de = "Die Anwendung wissenschaftlicher Erkenntnisse für praktische Zwecke"
  }
  
  note = {
    en = "Use this concept for all technology-related content"
  }
  
  notations = ["TECH", "T001"]
  
  # Hierarchical relationships
  broader = ["broader-concept-id"]
  narrower = ["narrower-concept-id-1", "narrower-concept-id-2"]
  related = ["related-concept-id"]
}