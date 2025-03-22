# BPJS Frista Face Recognition BOT

Rest Api untuk membuka aplikasi frista BPJS. Digunakan untuk implementasi APM (Anjungan Pendaftaran Mandiri) berbasis WEB.

## Build

untuk membuat exe

```bash
go build -o FingerprintBot.exe main.go
```

## Dokumentasi Rest API

Ping

```bash
curl --location --request POST 'http://localhost:3006/ping'
```

membuka aplikasi

```bash
curl --location 'http://localhost:3006/open' \
--header 'Content-Type: application/json' \
--data-raw '{
"username": "agung,arief",
"password": "Agung123#",
"no_bpjs": "0003498723872394"
}'
```

menutup aplikasi

```bash
curl --location 'http://localhost:3005/close' \
--header 'Content-Type: application/json' \
--data '{
}'
```