# TodoList Service

## Project layout

```
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
```

## Database 
---
<p align="center" width="100%">
    <img width="50%" src="database.png?raw=true"> 
</p>


## Requests Example
Get Tasks
```
curl --insecure --location --request GET 'https://localhost:11000/api/v1/task?page=1'
```
Get Task
```
curl --insecure --location --request GET 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483'
```
Delete Task
```
curl --insecure --location --request DELETE 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483'
```
Create Task
```
curl --insecure --location --request POST 'https://localhost:11000/api/v1/task' \
--header 'Content-Type: application/json' \
--data-raw '{
    "value": "task_1",
    "due_date": "2022-03-28T23:37:17.150Z"
}'
```
Update Task
```
curl --insecure --location --request PUT 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483' \
--header 'Content-Type: application/json' \
--data-raw '{
    "value": "task_1",
    "completed": true,
    "due_date": "2022-03-27T20:54:54.078Z"
}'
```
Task Status
```
curl --insecure --location --request PATCH 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/status' \
--header 'Content-Type: application/json' \
--data-raw '{
    "completed": true
}'
```
Get Comments
```
curl --insecure --location --request GET 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/comment'
```
Create Comment
```
curl --insecure --location --request POST 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/comment' \
--header 'Content-Type: application/json' \
--data-raw '{
    "comment": "Testing_123"
}'
```
Delete Comment
```
curl --insecure --location --request DELETE 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/comment/2d246b2a-447d-4c5e-bce6-4099aac5d049' \
--header 'Content-Type: application/json' \
--data-raw '{
    "comment": "Testing_123"
}'
```
Get Labels
```
curl --insecure --location --request GET 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/label'
```
Create Label
```
curl --insecure --location --request POST 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/label' \
--header 'Content-Type: application/json' \
--data-raw '{
    "label": "Testing_1234"
}'
```
Delete Label
```
curl --insecure --location --request DELETE 'https://localhost:11000/api/v1/task/aa54dc02-b5c4-4629-889e-ee64d3921483/label/91921fd2-f83f-4d26-ba0c-4c32334356c2'
```