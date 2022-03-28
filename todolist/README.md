## TodoList Service

### Project layout

├── Dockerfile.grpc
├── Dockerfile.grpc.dev
├── Dockerfile.http
├── Dockerfile.http.dev
├── Makefile
├── README.md
├── buf.gen.yaml
├── buf.lock
├── buf.yaml
├── certs
│   ├── README.md
│   ├── grpc
│   └── http
├── cmd
│   ├── grpc
│   └── http
├── go.mod
├── go.sum
├── internal
│   ├── grpc
│   │   ├── healthcheck
│   │   └── todolist
│   ├── model
│   └── repository
├── pkg
│   ├── mock
│   └── storage

├── proto
├── scripts
│   └── db
│       ├── 10-init.sql
│       └── migrations
├── third_party
├── tools

### Database 
---
<p align="center" width="100%">
    <img width="50%" src="service.png?raw=true"> 
</p>