# Mini-Scan

Mini-Scan provides a simple toy implementation of a pubsub-driven service scanning system. It consists of a scanner that simulates scanning live services and publishing the results to a pubsub topic. Readers then subscribe to the topic and store the results in a database. Readers can be horizontally scaled to any arbitrary number of instances. The database stores the latest response for a given service + ip + port combination. 

---

## Running the system locally

A docker-compose file is provided to run the system locally. A postgres database has been provided for the persistence layer, but this can be replaced with any other database by implementing the `Scanning` interface in `pkg/db/db_scan.go` and passing the database URL via the `-db-url` flag.

The system can be run locally by running:
```bash
docker compose up
```

This will start the postgres database, the pubsub emulator, and the scanner and reader services.

The scanner will publish scan results to the pubsub topic every second. The reader will subscribe to the topic and store the results in the database.

The system can be stopped by running:

```bash
docker compose down
```

## Scaling the system

The system can be scaled by running multiple reader instances.

```bash
docker compose scale reader=3
```

## Why Postgres?

Postgres was chosen because its conflict resolution capabilities allow for conditionally inserting/updating scan results in a single atomic SQL operation. In essence the problem is we want to insert a scan result if it doesn't exist, or update it if it does and *the updated scan result is more recent than the existing one.* With Postgres we can define our primary key to be a composite key of the service, ip, and port, and use the `ON CONFLICT` clause alongside a conditional `WHERE` clause to achieve atomic updates in one statement:

```sql
INSERT INTO scans AS s (ipv4_addr, port, service, resp, updated_at) -- write the results
VALUES (@ipv4_addr, @port, @service, @resp, to_timestamp(@updated_at))
ON CONFLICT (ipv4_addr, port, service) DO UPDATE SET --however if the scan result already exists
	resp = EXCLUDED.resp, -- update the service response
	updated_at = EXCLUDED.updated_at
WHERE EXCLUDED.updated_at > s.updated_at -- but only if the updated scan result is more recent than the existing one
```

If, for example, we had chosen SQLlite instead, then we would not have access to this same conflict resolution behavior. SQLlite does support an `INSERT OR REPLACE` statement, but it does not support conditionally updating the record. We would need to wrap multiple statements in a transaction and perform the logic within our application code to achieve the same result. 

## Testing

Manual verification was performed by running the system locally and checking the database for the latest scan result for a given service + ip + port combination. Automated tests were not implemented due to time constraints.