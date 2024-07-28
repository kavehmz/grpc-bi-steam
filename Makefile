.PHONY: all proto clean route relay connect notification_service recommendation_service

all: proto

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative hub/service_hub.proto

clean:
	rm -f hub/*.pb.go

route:
	go run route/main.go

relay:
	go run relay/main.go

connect_notification:
	go run connect/main.go -name notification -target http://localhost:8080

connect_recommendation:
	go run connect/main.go -name recommendation -target http://localhost:8081

notification_service:
	go run sample_external_services/notification_service/main.go

recommendation_service:
	go run sample_external_services/recommendation_service/main.go
