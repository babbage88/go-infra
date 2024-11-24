check-swagger:
	which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger: check-swagger
	swagger generate spec -o ./swagger.yaml --scan-models && swagger generate spec > swagger.json

embed-swagger:
	swagger generate spec -o ./embed/swagger.yaml --scan-models && swagger generate spec > ./embed/swagger.json

serve-swagger: check-swagger
	swagger serve -F=swagger swagger.yaml --no-open --port 4443
