# Quickmap - a simple in-memory key-value store

Quickmap is a simple and light weight in-memory key-value data store that can concurrently do operations on its underlying data structure. The type of key and value can be anything of int, float, string, and complex numbers.

For now, it has a server component. The server starts on port `5050` and it listens for `GET` requests at the `localhost:5050/api/v1/exec-command` endpoint.

## Request body structure

```json
{
	"command": "string"
}
```

## Response body structure

```json
{
	"code": 200,
	"value": null,
	"error": ""
}
```

The response structure is same for all commands. The `value` is `null` for some commands (e.g. `SET` command). The `code` is in sync with the response status.

One thing to note that the `value` can be of integer type, string type or float type.

## Available Commands

### SET

```
Usage: SET <key> <value> [EX <expiry>] [<condition>]

Set the Key Value pair with optional expiry duration and condition.

The supported types for <key> and <value> are string, float, int,
and complex numbers.

<expiry> is an optional integer value which denotes the duration in
second after which the key-value pair in the store will expire.

<condition> is an optional field used to control the SET behaviour.
Two supported values are NX and XX.
NX - Only set the key if it does not already exist
XX - Only set the key if it already exists
```

Example requests:

```bash
curl -X GET "127.0.0.1:5050/api/v1/exec-command" \
	-H "Content-type: application/json" \
	-d '{"command": "SET test 20.5 EX 60 NX"}'
```

Response:

```json
{
	"code":201,
	"value":null,
	"error":""
}
```

### GET

```
Usage: GET <key>

Get the value associated with the specified key.

The supported types for <key> are string, float, int, and
complex numbers.
```

Example Request:

```bash
curl -X GET "127.0.0.1:5050/api/v1/exec-command" \
	-H "Content-type: application/json" \
	-d '{"command": "GET test"}'
```

Response:

```json
{
	"code":200,
	"value":20.5,
	"error":""
}
```

### QPUSH

```
Usage: QPUSH <key> <value...>

Creates a queue if not already created with the <key> and
appends values to it.

<key> - Name of the queue to write to.
<value...> - Variadic input that receives multiple values separated by space.
```

Example Request:

```bash
curl -X GET "127.0.0.1:5050/api/v1/exec-command" \
	-H "Content-type: application/json" \
	-d '{"command": "QPUSH test 12 10 23 22 false 4.5"}
```

Example Request:

```json
{
	"code":201,
	"value":null,
	"error":""
}
```

### QPOP

```
Usage: QPOP <key>

Returns the last inserted value from the queue with the <key>.
```

Example Request:

```bash
curl -X GET "127.0.0.1:5050/api/v1/exec-command" \ 
	-H "Content-type: application/json" \
	-d '{"command": "QPOP test "}
```

Example Response:

```json
{
	"code":200,
	"value":4.5,
	"err": ""
}
```
### BQPOP

```
Usage: BQPOP <key> <timeout>

[Experimental] Blocking queue read operation that blocks the thread
until a value is read from the queue.

The command will fail if multiple clients try to read from the queue
at the same time.

<key> - Name of the queue to read from.
<timeout> - The duration in seconds to wait until a value is read from
the queue. The argument must be interpreted as a double value. A value
of 0 immediately returns a value from the queue without blocking.
```

Example Request:

```bash
curl -X GET "127.0.0.1:5050/api/v1/exec-command" \
	-H "Content-type: application/json" \
	-d '{"command": "BQPOP test 10.5"}
```

Request:

```json
{
	"code":200,
	"value":4.5,
	"err": ""
}
```

Note that the command name is case-sensitive and all the commands are required to be in capital letter. Otherwise it will return `unknown command` error.

## Future Improvements

Although this is a in-memory key-value store, it is far behind the standard stores in terms of efficieny, optimization and scaling (e.g. [redis](https://github.com/redis/redis), Dgraph's [Ristretto](https://github.com/dgraph-io/ristretto)). This is the reason I put the endpoint under `api/v1`. I will try to match this repository to the standard ones.

Currently the project doesn't have any unit tests. So I have to add tests also ;)