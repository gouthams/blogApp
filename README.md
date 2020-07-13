# Overview/Design
A sample blogging application backend api. 
A simple Go(Gin-Gonic) backend server to host blogging apis with the mongoDb as the persistence layer.
This hosts two CRUD apis blogUsers and blogPosts and runs on default http port 8080

Example:
post -> http://localhost:8080/blogUsers
get -> http://localhost:8080/blogPosts?userId=<60170ef7-2157-4c83-b0db-9efcf492f17d>&pageSize=10
   
## Assumptions:

1) To keep it simple, old version of mongoDb (3.4) is installed with in the Golang-alpine image. Look at DockerFile
 for the more details

2) For persistence as of now this application is tightly coupled with mongoDB. If need be a database abstraction layer
 can be added to use a different database or an external db service. 

3) blog-openapi.yaml file has the api spec definition. To keep the api models consistent open-api generator is used
 to generate the routers.go file. 
 NOTE: Upstream go-gin-server target for open-api generator has a bug. It does not set the proper binding in the
  models which is required for the data validation. So models are overridden with required bindings.
 
 4) Persistence volume is bind mounted to /tmp/db location. So this will be used for the storage. 

### Install and Build
Requires Golang installed. Please follow the instruction from here https://golang.org/doc/install
Requires Docker installed. https://docs.docker.com/get-docker/

This library is developed with go version 1.14.4

Download/clone the application code from from https://github.com/gouthams/blogApp

Needs access from github to resolve dependencies.

To build the docker container
```shell script
make dockerBuild
```

To run the docker container
```shell script
make dockerRun
```

To stop the running docker container and clean the residue
```shell script
make dockerStop
```

### Unit test
To build the docker test container
```shell script
make dockerBuildTest
```

To run the docker test container
```shell script
make dockerRunTest
```

To stop the running docker test container and clean the residue
```shell script
make dockerStopTest
```

To execute the unit test with the coverage profile local with a system installed with go, do the following
```shell script
cd github.com/gouthams/blogApp/server
go test -coverprofile cp.out
go tool cover -html=cp.out
```

### Future Consideration:
   1) Search string are limited to name,id and pageSize. To support paging to large queries with context further
    enhancements are required.
   
