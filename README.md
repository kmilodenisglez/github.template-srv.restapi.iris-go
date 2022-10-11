# 🛰 GitHub Template API server
REST API that allows clients to communicate with * (i.e. **dispatch controller**).

> **NOTE**: Drones app has been tested on **Ubuntu 18.04** and on **Windows 10 with WSL** and Golang 1.18 was used.

## Table of Contents

- [API specification](#api_spec)
- [Configuration file](#config_file)
- [Get Started](#get_started)
  * [Deployment ways (2 ways)](#deploy_ways)
    - [Docker way](#docker_way)
    - [Manual way](#manual_way)
- [Tech and packages](#tech)
- [Architecture](#arch)
- [SWAGGER](#swagger)
## ⚙️API specification <a name="api_spec"></a>

The **GitHub Template API server** provides the following Example API with communicating the **DB**:

| Tag           | Title                              | URL                                      | Query | Method |
| ------------- | ---------------------------------- | ---------------------------------------- | ----- | ---- |
| Auth          | user authentication (Using JWT)    | `/api/v1/auth`                           |   -   |`POST`|
| Auth          | user logout                        | `/api/v1/auth/logout`                    |   -   |`GET` |
| Auth          | get user authenticated             | `/api/v1/auth/user`                      |   -   |`GET` |
| Database      | Populate DB with fake data         | `/api/v1/database/populate`              |   -   |`POST`|
| Drones        | Get all drones or filters for State| `/api/v1/drones`                         |?state=|`GET` |
| Drones        | Registers or update a drone        | `/api/v1/drones`                         |   -   |`POST`|
| Drones        | Get a drone by serialNumber        | `/api/v1/drones/:serialNumber`           |   -   |`GET` |
| Medications   | Get medications                    | `/api/v1/medications`                    |   -   |`GET` |
| Medications   | Checking loaded items for a drone  | `/api/v1/medications/items/:serialNumber`|   -   |`GET` |
| Medications   | Load a drone with medication items | `/api/v1/medications/items/:serialNumber`|   -   |`POST`|

To see the API specifications in more detail, run the app and visit the swagger docs:

> http://localhost:7001/swagger/index.html

![swagger ui](/docs/images/swagger-ui.png)


## 🛠️️ Configuration file (conf.yaml) <a name="config_file"></a>
👉🏾 [The config file](/conf/conf.yaml)

|  Param      | Description       | default value   |
| ----------- | -----------|------------------------- |
| APIDocIP    | IP to expose the api (unused)  | 127.0.0.1
| DappPort    | app PORT              | 7001
| StoreDBPath | DB file location      | ./db/data.db
| CronEnabled | active the cron job   | true
| LogDBPath   | DB file event logs    | ./db/event_log.db
| EveryTime   | time interval (in seconds) that the cron task is executed | 300 seconds (every 5 minutes)

By default, **StoreDBPath** generates the database file in the /db folder at the root of the project.

The server exposes the `/api/v1/database/populate` POST endpoint to generate and repopulate the database whenever necessary.
## ⚡ Get Started <a name="get_started"></a>

Download the github.template-srv.restapi.iris.go project and move to root of project:
```bash
git clone https://github.com/kmilodenisglez/github.template-srv.restapi.iris.go.git && cd github.template-srv.restapi.iris.go 
```

### 🚀 Deployment ways (2 ways)  <a name="deploy_ways"></a>
You can start the server in 2 ways, the first is using **docker** and **docker-compose** and the second is **manually**
#### 📦 Docker way <a name="docker_way"></a>
You will need docker and docker-compose in your system.

To builds Docker image from  Dockerfile, run:
```bash
docker build --no-cache --force-rm --tag app_restapi .
```
Use docker-compose to start the container:
```bash
docker-compose up
```

#### 🔧 Manual way  <a name="manual_way"></a>

Run:
```bash
go mod download
go mod vendor
```

If you make changes to the Endpoint you must generate Swagger API Spec:
 
![swagger doc](/docs/swagger.md)

Build:
```bash
go build
```

#### 🌍 Environment variables
The environment variable is exported with the location of the server configuration file.

If you have 🐧Linux or 🍎Dash, run:
```bash
export SERVER_CONFIG=$PWD/conf/conf.yaml
```
but if it is in the windows cmd, then run:
```bash
set SERVER_CONFIG=%cd%/conf/conf.yaml
```
#### 🏃🏽‍♂️ Start the server
Before it is recommended that you read more about the server configuration file in the section 👉🏾  .

Run the server:
```bash
./restapi.app
```

and visit the swagger docs:

> http://localhost:7001/swagger/index.html

The first endpoint to execute must be /api/v1/database/populate [POST], to populate the database. That endpoint does not need authentication.

![swagger ui](/docs/images/populate_endpoint.png)

You can then authenticate and test the remaining endpoints.

### 🧪 Unit or End-To-End Testing
Run:
```bash
go test -v
```

## 🔨 Tech and packages <a name="tech"></a>
* [Iris Web Framework](https://github.com/kataras/iris)
* [validator/v10](https://github.com/go-playground/validator)
* [Buntdb](https://github.com/tidwall/buntdb)
* [govalidator](https://github.com/asaskevich/govalidator)
* [gocron](https://github.com/go-co-op/gocron)
* [swag](https://github.com/swaggo/swag)
* [Docker](https://docs.docker.com)
* [docker-compose](https://docs.docker.com/compose/)

## 📐 Architecture <a name="arch"></a>
This project has 3 layer :

- Controller Layer (Presentation)
- Service Layer (Business)
- Repository Layer (Persistence)


Tag | Path | Layer |
--- | ---- | ----- |
Auth     | [end_auth.go](/api/endpoints/end_auth.go) | Controller | 
Drones   | [end_drones.go](/api/endpoints/end_drones.go) |  Controller |
EventLog | [end_eventlog.go](/api/endpoints/end_eventlog.go) |  Controller |
 |  |  |
Auth     | [svc_authentication.go](/service/auth/svc_authentication.go) | Service | 
Drones   | [svc_drones.go](/service/svc_drones.go) |  Service |
EventLog | [svc_eventlog.go](/service/cron/svc_eventlog.go) |  Service |
 |  |  |
Auth     | [repo_drones.go](/repo/db/repo_drones.go) | Repository | 
Drones   | [repo_drones.go](/repo/db/repo_drones.go) |  Repository |
EventLog | [repo_eventlog.go](/repo/db/repo_eventlog.go) |  Repository |

## 📐 Swagger <a name="swagger"></a>
Read ![swagger doc](/docs/swagger.md)