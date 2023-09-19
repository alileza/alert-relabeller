test:
	@go test ./... 

tomato-build:
	@docker-compose -f ./tomato/docker-compose.yaml build

tomato-test:
	@docker-compose -f ./tomato/docker-compose.yaml up --abort-on-container-exit