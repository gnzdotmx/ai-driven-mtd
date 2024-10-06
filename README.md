# AI-Driven Moving Target Defense System

This repository contains an AI-driven Moving Target Defense (MTD) system that uses Elasticsearch and Ollama3 model to make decisions on how to change configurations on docker containers. The changes are on the Operating system, code language the app was built, and the content type format (json, yaml).

The MTD system aims to improve security, reduce costs, optimize performance, etc., by making dynamic changes to the configuration of a system based on the current metrics and previous decisions made by Subject Matter Experts (SMEs).

![alt Architecture](./imgs/architecture.gif)

## Components

The system consists of the following components:

- **app_golang**: The web service written in Golang consumed by end-users.
- **app_python**: The web service written in Python consumed by end-users.
- **client**: Simulates a client application that sends requests to the web services for metrics collection. (Testing purposes)
- **config**: Configuration files for the MTD system.
    - **client_config.json**: Configuration files for the client application. Servers, request interval, etc.
    - **config.json**: Configuration settings for the MTD system. Supported ports, OSes, formats, languages, etc.
	- **knowledge.json**: Knowledge base for the MTD system. Previous decisions, recommendations, etc., taken by SMEs.
    - **metrics.json**: Current metrics configuration for the MTD system. Thresholds, weights, etc. This should be collected by another system, so far it is manually set.
- **docker**: Dockerfiles for setting up supported OSs for movements, and Ollama + Elasticsearch services.
- **mtd**: The main package that contains the core logic for the MTD system. Strategies, decision-making, etc.
- **ollama**: Code in golang to interact with Ollama API.
- **scripts**: Helper scripts for setting environment variables and run services.

## Examples 
`run` output example
```bash
❯ make run
go run main.go
2024/10/06 00:45:33 Connected to Elasticsearch
2024/10/06 00:45:33 Available configurations:
                {Ports:[8080 8081 8082 8083] OSes:[golang python ubuntu] Formats:[json yaml text] Languages:[golang python]}
2024/10/06 00:45:33 
        Searching on Elasticsearch for:
                response time: 200.000000
                error rate: 0.020000
                vulnerability count: 49
                intrusion attempts: 61

2024/10/06 00:45:33 Best matches on Elasticsearch:
                {PolicyName:Security Tightening Criteria:{ResponseTimeMs:200 ErrorRate:0.03 VulnerabilityCount:50 IntrusionAttempts:60} RecommendedActions:{SwitchLanguage:golang SwitchFormat:yaml SwitchOS:golang RotateIP:true}}
                {PolicyName:Security Hardening Criteria:{ResponseTimeMs:100 ErrorRate:0.02 VulnerabilityCount:30 IntrusionAttempts:100} RecommendedActions:{SwitchLanguage:python SwitchFormat:yaml SwitchOS:golang RotateIP:true}}
                {PolicyName:Cost Cutting Criteria:{ResponseTimeMs:100 ErrorRate:0.02 VulnerabilityCount:17 IntrusionAttempts:35} RecommendedActions:{SwitchLanguage:python SwitchFormat:text SwitchOS:golang RotateIP:true}}
2024/10/06 00:46:55 
                Ollama> { "SwitchLanguage": "python", "SwitchOS": "ubuntu", "SwitchFormat": "json", "SwitchPort": "443", "RotateIP": "true"}
2024/10/06 00:46:56 MTD changes applied: PORT= OS=ubuntu, Format=json, Language=python
```


Once it finishes, test communication and check received headers by using `curl`. As you can see the server answered with `Content-Type` = `json`, `X-Server-Language` = `python` which is the language the app was built, and `X-Server-OS` = `ubuntu` as Ollama recommended. The port was not changed as these lines were commented in the code. Keep reading to know why.
Finally, you can see the server answered with a JSON format output, so the client could process the output based on the headers.

```bash
❯ curl -v http://127.0.0.1:8080/
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080
> GET / HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/8.7.1
> Accept: */*
> 
* Request completely sent off
< HTTP/1.1 200 OK
< Server: Werkzeug/3.0.4 Python/3.8.10
< Date: Sat, 05 Oct 2024 16:53:48 GMT
< Content-Type: application/json
< Content-Length: 28
< X-Server-Language: python
< X-Server-OS: ubuntu
< Connection: close
< 
* Closing connection
{"message": "Hello, World!"}%  
```

# Usage
### Requirements
- Docker
- Docker compose
- curl
- python3
- jq
- shuf


### Setup the environment
Setup Ollama and Elasticsearch services:
```bash
make start
```
Download the Ollama3 model:
```bash
docker exec -it ollama ollama pull llama3:latest
```

### Run the MTD system
Run it every time you want to make a change to the system. 
```bash
make run
```
### Stop the environment
Stop the Ollama and Elasticsearch services:
```bash
make stop
```

### Clean up the environment
Stop and remove containers, ollama, and elasticsearch folders:
```bash
make clean
```

# MTD Strategies
The MTD system currently supports the following strategies:

- **Round-Robin**: Cycles through the available configurations in a round-robin fashion.
- **Random**: Randomly selects a configuration from the available configurations.
- **Weighted**: Uses a weighted AI-Driven decision-making algorithm to select the best configuration based on the current metrics and previous decisions.

## Weighted Strategy
The weighted strategy uses a weighted decision-making algorithm to select the best configuration based on the current metrics and previous decisions.
The decision is taken based on:
- Retrieved knowledge from Elasticsearch, which contains previous decisions made by SMEs.
- Ask Ollama for a recommendation based on the current metrics and the retrieved knowledge.

The configuration is as follows:
- The weights that describe the current system's metrics are stored in `config/metrics.json`.
- The knowledge database is stored in `config/knowledge.json`. It describes the criteria used by the SMEs to make decisions, also contains the decisions made by them and are labeled as `recommended actions`.
- The available movements are stored in the `config/config.json` file. This file describes the available configurations for the system. Ports, OSes, Formats, Languages, etc.

# Debugging
Check environment variables set to the running container. You should see RESPONSE_FORMAT, RESPONSE_OS, and RESPONSE_LANGUAGE.
```bash
docker exec -it docker-app_ubuntu_python-1 bash -c "env"
```

Test communication and review received headers from the deployed application.
```bash
curl -v http://localhost:8080/
```

Check IPs assigned to docker containers.
```bash
docker network ls
docker network inspect docker_mtd_network
```

Check running containers
```bash
docker ps
```

# Future work
For future work and testing you can explore the code and here are some initials interesting points

- In `main.go` you can find commented code which enters in a loop to keep `moving` the environment.
- By default, in `mtd/elastic.go` the code pulls maximum 5 matches from elasticsearch. Change it if you want to give more examples to Ollama and get better results.
- You can add more LLMs like ChatGPT or Gemini to have better results if you machine does not have enough resources to get good results.
- The port is not being changed yet as we need to design a way for the client to get such port, or figure out an application where changing the port is applicable.