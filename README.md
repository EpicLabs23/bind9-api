# bind9-api

REST API with Go to manage BIND9 DNS Server from anywhere

## Postman Collection:

https://api.postman.com/collections/915142-834e437b-a7fa-44b4-98f7-b4f56fa7ad34?access_key=PMAT-01K3DZ1BG19Q9VQZQMSRMX4ZSQ

## Installation

Direct from binary

```bash
curl -sSL https://raw.githubusercontent.com/EpicLabs23/bind9-api/main/install.sh | bash -s 0.1.0

```

Debian/Ubuntu:

```bash
wget https://github.com/EpicLabs23/bind9-api/releases/download/v<version>/bind9-api_<version>_linux_amd64.deb
sudo dpkg -i bind9-api_<version>_linux_amd64.deb
sudo systemctl enable bind9-api
sudo systemctl start bind9-api
```

## CentOS/Fedora:

```bash
wget https://github.com/EpicLabs23/bind9-api/releases/download/v<version>/bind9-api_<version>_linux_amd64.rpm
sudo rpm -i bind9-api_<version>_linux_amd64.rpm
sudo systemctl enable bind9-api
sudo systemctl start bind9-api
```

## Manual Tarball

```bash
wget https://github.com/EpicLabs23/bind9-api/releases/download/v<version>/bind9-api_<version>_linux_amd64.tar.gz
tar -xzf bind9-api_<version>_linux_amd64.tar.gz
sudo mv bind9-api /usr/local/bin/
sudo mkdir -p /etc/bind9-api
sudo mv config.yml /etc/bind9-api/
sudo cp packaging/systemd/bind9-api.service /etc/systemd/system/
sudo systemctl enable bind9-api
sudo systemctl start bind9-api
```

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
