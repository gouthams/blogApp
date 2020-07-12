# Overview/Design
A sample blogging application backend api. 
A simple Go(Gin-Gonic) backend server to host blogging apis with the mongoDb as the persistence layer. 
   
## Assumptions:

1) To keep it simple old version of mongoDb (3.4) is installed with in the Golang-alpine image. Look at DockerFile
 for the more details

2) blog-openapi.yaml file has the api spec definition. To keep the api models consistent open-api generator is used
 to generate the routers.go file. 
 NOTE: Upstream go-gin-server target for open-api generator has a bug. It does not set the proper binding in the
  models which is required for the data validation. So models are overridden with required bindings. 
     
3) For logging with log levels support third party logrus logger has been used https://github.com/Sirupsen/logrus


### Install and Build
Requires Golang installed. Please follow the instruction from here https://golang.org/doc/install

This library is developed with go version 1.14.4

Download the library from https://github.com/gouthams/

Need access from github to resolve dependencies.

To build the docker container
```shell script
make dockerBuild
```

To run the docker container
```shell script
make dockerRun
```

To stop the running docker container
```shell script
make dockerStop
```

To execute the unit test with the coverage profile, do the following
```shell script
cd server
go test -coverprofile cp.out
go tool cover -html=cp.out
```

### Unit test
For assertion in unit test, this library is used https://github.com/stretchr/testify. 
This go module dependency should be resolved during the build time.  

### Future Consideration:
   1) Search string are limited to name,id and pageSize. To support paging to large queries with context further
    enhancements are required.
   2) 
   
