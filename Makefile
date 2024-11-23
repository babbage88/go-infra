check-swagger:
	which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger: check-swagger
	swagger generate spec -o ./swagger.yaml --scan-models --exclude-deps

serve-swagger: check-swagger
	swagger serve -F=swagger swagger.yaml