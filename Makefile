OPEN_API_VERSION=v4.3.1
GENERATOR=go-gin-server
TEMP_FOLDER=openapitemp

g: generate
generate:
	docker run --rm -v `pwd`:/local \
	  openapitools/openapi-generator-cli:${OPEN_API_VERSION} generate \
	  --input-spec /local/blog-openapi.yaml \
	  --generator-name ${GENERATOR} \
	  --additional-properties=packageName=restimpl \
	  --output /local/${TEMP_FOLDER}
	cp ./${TEMP_FOLDER}/go/routers.go ./server/restimpl/routers.go
# Upstream bug in openapi go-gin generator - wont generate bindings
#	mkdir -p ./server/model
#	cp ./${TEMP_FOLDER}/go/model_*.go ./server/model/
	rm -r ./${TEMP_FOLDER}

# clean up all the local files produced by generating, building and testing
c: clean
clean:
	-rm -r ./${TEMP_FOLDER}
	-rm  server/restimpl/routers.go
	-rm ./blogApp

f: format
format:
	go fmt ./...

install: format
	go install -v ./...

dockerBuild: clean generate
	docker build -t blogapp:latest -f Dockerfile .

dockerRun:
	docker run -p 8080:8080 --name blogAppContainer -v /tmp/db:/data/db blogapp:latest

dockerStop:
	docker stop $(shell docker ps -aqf "name=blogAppContainer")
	docker rm $(shell docker ps -aqf "name=blogAppContainer")

# just runs a go build on local files
# NOTE: Need a local mongoDD running to host the server
localRun: format
	go build -o blogApp ./server/
	./blogApp

localBuild: clean generate install

cmod: clean-mod
clean-mod:
	GO111MODULE=on go clean -modcache