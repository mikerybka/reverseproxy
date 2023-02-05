docker-compose example:

```
---
version: "2.1"
services:
  reverseproxy:
    image: mikerybka/reverseproxy:latest
    environment:
      - EMAIL=name@example.org
    volumes:
      - /path/to/config:/etc/reverseproxy
      - /path/to/certs:/etc/ssl/certs
      - /path/to/logs:/var/log/reverseproxy
    ports:
      - 80:80
      - 443:443
    restart: unless-stopped
```
