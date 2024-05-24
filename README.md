# Run

Prerequisites:
- docker >= 20.10 or docker-compose >= 1.27.0

```bash
docker compose --compatibility -f docker-compose.yaml build
docker compose --compatibility -f docker-compose.yaml up -d
```

# Quick demo

```bash
id=$(curl -sX POST http://localhost:8080/user -d '{"password":"12345678","first_name":"John","last_name":"Doe","birthdate":"1992-05-01","gender":"male","interests":"Work","city":"Moscow"}' | jq -r .id)
token=$(curl -sX POST http://localhost:8080/login -d '{"id": '$id', "password": "12345678"}' | jq -r .token | cut -d " " -f 3)
curl -s http://localhost:8080/user/$id -H "Authorization: Bearer $token"; echo
```

# Teardown

```bash
docker compose --compatibility -f docker-compose.yaml down -v
```
