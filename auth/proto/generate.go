package proto

//go:generate protoc ./auth.proto --go_out=./generated-source --go-grpc_out=./generated-source

//go:generate mockgen -source=./generated-source/auth_grpc.pb.go -destination ./generated-source/auth_grpc.pb_mock.go -package auth_grpc_api
