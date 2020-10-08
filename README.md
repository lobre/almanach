# Almanach

Web app to manage events in a music band

## Database

To create a demo postgres database:
```
docker run --name almanach-db -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres
```

To browse via the psql command line tool:
```
docker exec -ti almanach-db psql -U postgres
```

To create Almanach tables:
```
docker exec -i almanach-db psql -U postgres < schema.sql
```

## Certs

If you run the server with tls support on a local domain, you will need to generate self signed certificates in the `certs` directory.

```
mkdir -p certs
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout certs/localhost.key -out certs/localhost.crt \
    -subj "/C=US/ST=Oregon/L=Portland/O=Company Name/OU=Org/CN=localhost"
```
