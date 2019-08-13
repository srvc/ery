// Code generated by github.com/izumin5210/gex. DO NOT EDIT.

// +build tools

package tools

// tool dependencies
import (
	_ "github.com/golang/protobuf/protoc-gen-go"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger"
	_ "github.com/izumin5210/grapi/cmd/grapi"
	_ "github.com/izumin5210/grapi/cmd/grapi-gen-command"
	_ "github.com/izumin5210/grapi/cmd/grapi-gen-scaffold-service"
	_ "github.com/izumin5210/grapi/cmd/grapi-gen-service"
	_ "github.com/izumin5210/grapi/cmd/grapi-gen-type"
)

// If you want to use tools, please run the following command:
//  go generate ./tools.go
//
//go:generate go build -v -o=./bin/protoc-gen-go github.com/golang/protobuf/protoc-gen-go
//go:generate go build -v -o=./bin/protoc-gen-grpc-gateway github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
//go:generate go build -v -o=./bin/protoc-gen-swagger github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
//go:generate go build -v -o=./bin/grapi github.com/izumin5210/grapi/cmd/grapi
//go:generate go build -v -o=./bin/grapi-gen-command github.com/izumin5210/grapi/cmd/grapi-gen-command
//go:generate go build -v -o=./bin/grapi-gen-scaffold-service github.com/izumin5210/grapi/cmd/grapi-gen-scaffold-service
//go:generate go build -v -o=./bin/grapi-gen-service github.com/izumin5210/grapi/cmd/grapi-gen-service
//go:generate go build -v -o=./bin/grapi-gen-type github.com/izumin5210/grapi/cmd/grapi-gen-type
