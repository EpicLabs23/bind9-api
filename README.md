# bind9-api

REST API with Go to manage BIND9 DNS Server from anywhere

Postman Collection:
https://api.postman.com/collections/915142-834e437b-a7fa-44b4-98f7-b4f56fa7ad34?access_key=PMAT-01K3DZ1BG19Q9VQZQMSRMX4ZSQ

## Devlopment Environment Setup

1. Clone the repository

```bash
git clone https://github.com/EpicLabs23/bind9-api.git
```

2. Change directory

```bash
cd bind9-api
```

3. Run

```bash
docker compose up -d
```

4. Start Dev aapplication

```bash
cd src
go mod tidy
docker exec -it bind9-api bash
cd /bind9-api/src
air
```
