# Almanach

Web app to manage events in a music band

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
