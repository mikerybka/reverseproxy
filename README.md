# reverseproxy

`reverseproxy` implements HTTPS termination and reverse proxying to localhost ports.

## Install

## Configure

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
