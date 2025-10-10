# Warning: EVERY LINE OF CODE IS WRITTEN BY HUMAN

## How to run

### Minikube(prerequisite)
```minikube start```


### Postgres for local dev
```docker-compose up -d```

### Run scripts/setup-keycloak.sh if not ran inside compose
```sh scripts/setup-keycloak.sh```


### Run server
```make setup```
```make dev```
