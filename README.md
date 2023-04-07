# batproxy
A HTTP reverse proxy, but connecting to the backend through SSH.

## Develop Guidance
* Create config file, example:
```yaml
# batproxy listen address
listen: :18880
proxies:
    # id is used as index
  - id: 0
    # uuid is unique proxy rule id
    uuid: localhost
    identity_file: .ssh/id_ecdsa
    #password: ssh_password
    user: user1
    host: ssh_host:22
    # reverse proxy http server ip
    node: 192.168.0.1
    # reverse proxy http server port
    port: 18880
```
