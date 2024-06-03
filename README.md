# Notifications Building Block
The Notifications Building Block manages user notifications for the Rokwire platform.

## Documentation
The functionality provided by this application is documented in the [Wiki](https://github.com/rokwire/notifications-building-block/wiki).

The API documentation is available here: https://api.rokwire.illinois.edu/notifications/api/doc/ui/index.html

## Set Up

### Prerequisites
MongoDB v4.2.2+

Go v1.22+

### Environment variables
The following Environment variables are supported. The service will not start unless those marked as Required are supplied.

Name|Format|Required|Description
---|---|---|---
HOST | < url > | yes | URL where this application is being hosted
PORT | < int > | yes | Port to be used by this application
MONGO_AUTH | <mongodb://USER:PASSWORD@HOST:PORT/DATABASE NAME> | yes | MongoDB authentication string. The user must have read/write privileges.
MONGO_DATABASE | < string > | yes | MongoDB database name
MONGO_TIMEOUT | < int > | no | MongoDB timeout in milliseconds. Defaults to 500.
INTERNAL_API_KEY | < string > | yes | Internal API key for invocation by other BBs
CORE_AUTH_PRIVATE_KEY | < string (PEM) > | yes | Private key for communicating with Core
CORE_BB_HOST | < url > | yes | Core BB host URL
NOTIFICATIONS_SERVICE_URL | < url > | yes | Notifications BB base URL
SMTP_EMAIL_FROM | < email > | yes | SMTP email from
SMTP_HOST | < string > | yes | SMTP host
SMTP_USER | < string > | yes | SMTP username
SMTP_PASSWORD | < string > | yes | SMTP password
SMTP_PORT | < int > | yes | SMTP port (Example 587)
NOTIFICATIONS_MULTI_TENANCY_ORG_ID | < string > | yes | Organization id for preparing the currently existing data to meet the multi-tenancy requirments(temporary field)
NOTIFICATIONS_MULTI_TENANCY_APP_ID | < string > | yes | Application id for preparing the currently existing data to meet the multi-tenancy requirments(temporary field)
AIRSHIP_HOST | < string > | yes | Airship host


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

## Contributing
If you would like to contribute to this project, please be sure to read the [Contributing Guidelines](CONTRIBUTING.md), [Code of Conduct](CODE_OF_CONDUCT.md), and [Conventions](CONVENTIONS.md) before beginning.

### Secret Detection
This repository is configured with a [pre-commit](https://pre-commit.com/) hook that runs [Yelp's Detect Secrets](https://github.com/Yelp/detect-secrets). If you intend to contribute directly to this repository, you must install pre-commit on your local machine to ensure that no secrets are pushed accidentally.

```
# Install software 
$ git pull  # Pull in pre-commit configuration & baseline 
$ pip install pre-commit 
$ pre-commit install