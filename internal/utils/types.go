package utils

var baseContentTypes = []string{
	"Symbol",
	"Text",
	"RichText",
	"Integer",
	"Number",
	"Date",
	"Boolean",
	"Object",
	"Location",
	"Array",
	"Link",
	"ResourceLink",
}

func GetContentTypes() []string {
	return append(baseContentTypes, "ResourceLink")
}

func GetAppFieldTypes() []string {
	return baseContentTypes
}

func GetLinkTypes() []string {
	return []string{"Asset", "Entry"}
}
