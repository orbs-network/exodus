# Exodus

Exodus is a virtual chain migration tool

## Running locally

```
docker run -d -p 5432:5432 -e POSTGRES_DB=exodus -e POSTGRES_USER=username -e POSTGRES_PASSWORD=password postgres
```

Update `config.json` file with relevant information.

To import transactions from `blocks` file into a local database, run

```
exodus import
```

To start migration process to a new vchain, run

```
exodus migrate
```
