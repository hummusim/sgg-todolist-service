## TODO LIST SERVICE Sensys Gatso Group

### Docs
---
- [Todolist](todolist/README.md) 
  
### How to run?
---
Requirements:
 - docker
 - docker-compose

```bash
./run-dev.sh
```

### Areas for improvement.
---
- A Code Review, I may have overlooked many things.
- Request validations.
- Integration test and more unit and detailed testing. For example, currently, I only tested if specific errors were returned. But I didn't make any comparison of objects.

### Architecture 
---
<p align="center" width="100%">
    <img width="50%" src="service.png?raw=true"> 
</p>

### Resources
- Inspired by https://github.com/grpc-ecosystem/grpc-gateway