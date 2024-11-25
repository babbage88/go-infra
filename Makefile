DOCKER_HUB:=jtrahan88/goinfra:

check-swagger:
	which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger: check-swagger
	swagger generate spec -o ./swagger.yaml --scan-models && swagger generate spec -o swagger.json --scan-models

dev-swagger: check-swagger
	swagger generate spec -o ./dev-swagger.yaml --scan-models && swagger generate spec -o dev-swagger.json --scan-models
	swagger mixin spec/swagger.dev.json dev-swagger.json --output swagger.json --format=json
	swagger mixin spec/swagger.dev.yaml dev-swagger.yaml --output swagger.yaml --format=yaml
	rm dev-swagger.json && rm dev-swagger.yaml

local-swagger: check-swagger
	swagger generate spec -o ./local-swagger.yaml --scan-models && swagger generate spec --scan-models -o local-swagger.json --scan-models
	swagger mixin spec/swagger.local.json local-swagger.json --output swagger.json --format=json
	swagger mixin spec/swagger.local.yaml local-swagger.yaml --output swagger.yaml --format=yaml
	rm local-swagger.json && rm local-swagger.yaml

run-local: local-swagger
	go run .

embed-swagger:
	swagger generate spec -o ./embed/swagger.yaml --scan-models && swagger generate spec > ./embed/swagger.json

serve-swagger: check-swagger
	swagger serve -F=swagger swagger.yaml --no-open --port 4443

buildandpush: dev-swagger
	docker buildx build --platform linux/amd64,linux/arm64 -t $(DOCKER_HUB)$(tag) . --push

deploydev:
	kubectl apply -f deployment/kubernetes/go-infra.yaml
	kubectl rollout restart deployment go-infra
