<!-- generated-from:0aed9be8805920c7a2e4f7e2f849ac573d82d0be60a0c6c34a13069a702255ff DO NOT REMOVE, DO UPDATE -->
# ACH Test Harness
**[Purpose](README.md)** | **[Configuration](CONFIGURATION.md)** | **Running** | **[Client](../pkg/client/README.md)**

--- 

## Running

### Getting started

More tutorials to come on how to use this as other pieces required to handle authorization are in place!

- [Using docker-compose](#local-development)
- [Using our Docker image](#docker-image)

No configuration is required to serve on `:8200` and metrics at `:8201/metrics` in Prometheus format.

### Docker image

You can download [our docker image `moov/ach-test-harness`](https://hub.docker.com/r/moov/ach-test-harness/) from Docker Hub or use this repository. 

### Local development

```
make run
```

---
**[Next - Client](../pkg/client/README.md)**