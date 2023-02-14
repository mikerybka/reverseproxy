# reverseproxy

`reverseproxy` implements HTTPS termination and reverse proxying to localhost ports.

## Docker example

```
docker run -it \
  -e EMAIL=you+letsencrypt@example.org # optional \
  -v path/to/app/data:/etc/reverseproxy
```

## Configuration

### Environment

#### `EMAIL` (optional)

The email to share with Let's Encrypt.

### Files

#### `/etc/reverseproxy/hosts.json`

A JSON file mapping hostnames to ports

##### Example
```json
{
  "example.org": "3000",
  "api.example.org": "8000"
}
```

## App Data

### `/etc/reverseproxy/certs`

SSL certificates from Let's Encrypt.

### `/etc/reverseproxy/logs`

Request logs.
Each request is encoded in a human-readable JSON file named after its [UnixNano](https://pkg.go.dev/time#Time.UnixNano) timestamp.
The request structure is defined by [web.Request](pkg/web/request.go).

<!-- Note: These logs can take up a non-trivial amount of space over time.
I recommend cleaning up this directory periodically.
Personally, I run a background task to process and clean up logs every 30 minutes. -->
