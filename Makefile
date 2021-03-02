SHELL := /bin/bash

MONGO_CONTAINER:="mongo_go_test"

test:
	go test ./pagingQuery_test.go

mongodb:
ifeq ("$(MONGO_CONTAINER)", "mongo_go_test")
	docker start mongo_go_test
else
	docker run --name mongo_go_test \
	-e MONGO_INITDB_ROOT_USERNAME=root \
	-e MONGO_INITDB_ROOT_PASSWORD=example \
	-v ${PWD}/mongo-data:/data/db \
	-p 27017:27017 \
	-d mongo:latest
endif

