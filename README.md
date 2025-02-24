# see-containers

## What it does
Basically trying to see the containers with Web UI.

## Planned Features
- [ ] Start and stop containers
- [ ] Show container's log (maybe real time, idk)

## Requirement
- Go
- Docker Compose

## Installation
### A. With Docker as Container
1. Modify docker_compose.yaml as your need (mostly port so it would conflict with other containers)
2. Build the image and Run Docker compose
```
docker compose build
docker compose up -d
```
or
```
docker compose up -d --build
```
### B. Go CLI for development
1. Clone the project
```
git clone https://github.com/rizkidn17/see-containers.git
```
2. Export the HOST IP and PORT in terminal (reset each session). Example :
```
export HOST_IP=172.31.0.2
export PORT=8080
```
3. Install dependencies
```
go mod download
go mod tidy
```
4. Run the app
```
go run ./cmd/see-containers
```
