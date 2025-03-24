.PHONY: build, test-unit, interactive, testacc

build:
	go build

test-unit: build
	sudo docker run \
		-e CONTENTFUL_MANAGEMENT_TOKEN=${CONTENTFUL_MANAGEMENT_TOKEN} \
		-e CONTENTFUL_ORGANIZATION_ID=${CONTENTFUL_ORGANIZATION_ID} \
		-e SPACE_ID=${SPACE_ID} \
		-e "TF_ACC=true" \
		terraform-provider-contentful \
		go test ./... -v

interactive:
	sudo -S docker run -it \
		-v $(shell pwd):/go/src/github.com/labd/terraform-provider-contentful \
		-e CONTENTFUL_MANAGEMENT_TOKEN=${CONTENTFUL_MANAGEMENT_TOKEN} \
        -e CONTENTFUL_ORGANIZATION_ID=${CONTENTFUL_ORGANIZATION_ID} \
        -e SPACE_ID=${SPACE_ID} \
		terraform-provider-contentful \
		bash

testacc:
	TF_ACC=1 go test -v ./...
