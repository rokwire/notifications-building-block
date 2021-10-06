# Notifications building block

Go project to provide rest service for rokwire notifications building block.

## Set Up

### Prerequisites

MongoDB v4.2.2+

Go v1.16+

### Environment variables
The following Environment variables are supported. The service will not start unless those marked as Required are supplied.

Name|Value|Required|Description
---|---|---|---
HOST | < value > | yes | Host name
PORT | < value > | yes | The port number of the listening port
MONGO_AUTH | <mongodb://USER:PASSWORD@HOST:PORT/DATABASE NAME> | yes | MongoDB authentication string. The user must have read/write privileges.
MONGO_DATABASE | < value > | yes | MongoDB database name
MONGO_TIMEOUT | < value > | no | MongoDB timeout in milliseconds. Set default value(500 milliseconds) if omitted
FIREBASE_PROJECT_ID | < value > | yes | Firebase project ID
FIREBASE_AUTH | < value > | yes | Firebase authentication JSON string
INTERNAL_API_KEY | < value > | yes | Internal API key for invocation of other BBs
CORE_AUTH_PRIVATE_KEY | < value > | yes | Private key for comunicating with Core
CORE_SERVICE_REG_LOADER_URL | < value > | yes | Core service reg loader url
NOTIFICATIONS_SERVICE_URL | < value > | yes | Notification base URL


### Run Application

#### Run locally without Docker

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Make the project  
```
$ make
...
▶ building executable(s)… 1.9.0 2020-08-13T10:00:00+0300
```

4. Run the executable
```
$ ./bin/notifications
```

#### Run locally as Docker container

1. Clone the repo (outside GOPATH)

2. Open the terminal and go to the root folder
  
3. Create Docker image  
```
docker build -t notifications .
```
4. Run as Docker container
```
docker-compose up
```

#### Tools

##### Run tests
```
$ make tests
```

##### Run code coverage tests
```
$ make cover
```

##### Run golint
```
$ make lint
```

##### Run gofmt to check formatting on all source files
```
$ make checkfmt
```

##### Run gofmt to fix formatting on all source files
```
$ make fixfmt
```

##### Cleanup everything
```
$ make clean
```

##### Run help
```
$ make help
```

##### Generate Swagger docs
```
$ make swagger
```

### Test Application APIs

Verify the service is running as calling the get version API.

#### Call get version API

curl -X GET -i https://api-dev.rokwire.illinois.edu/notifications/api/version

Response
```
0.1.2
```

## Documentation

The documentation is placed here - https://api-dev.rokwire.illinois.edu/docs/

Alternativelly the documentation is served by the service on the following url - https://api-dev.rokwire.illinois.edu/notifications/doc/ui/
