
.PHONY: lint-openapi-spec
lint-openapi-spec:
	# redocly/cli:1.26.0
	docker run --rm -v $(PWD):/spec redocly/cli@sha256:f2239c9097b09fa8ede54bb9b09d5ac6bba781bfe50606161f9adfce5fdc7b11 lint --config=.redocly.yaml $(APIS)
