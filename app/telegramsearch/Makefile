
.PHONY: lint-openapi-spec
lint-openapi-spec:
  # example-app@v1 originates from /.redocly.yaml in the root folder.
  # one should register their API in the config file above first, for this command to work.
  # APIS is a space separated list of OpenAPI specs registered in the /.redocly.yaml.
	cd ../.. && make lint-openapi-spec APIS="telegram-search@v1"
