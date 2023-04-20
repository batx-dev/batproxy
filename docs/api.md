# Manager Proxy API

## Use by unix socket

* `curl` add `--unix-socket` option

## Create a reverse proxy rule

```shell
$ curl -X POST --header "Content-Type: application/json" \
    http://localhost:18888/api/v1beta1/proxies -d \
    '{ 
        "user": "<cluster_login_name>", 
        "host": "<cluster_login_host>", 
        "password": "<cluster_login_password>", 
        "node": "<node_name_or_node_ip>", 
        "port": <port> 
    }'

    { 
        "proxy_id": "<proxy_id>"
        "user": "<cluster_login_name>", 
        "host": "<cluster_login_host>", 
        "password": "<cluster_login_password>", 
        "node": "<node_name_or_node_ip>", 
        "port": <port> 
    }
    
# Example
$ curl -X POST --header "Content-Type: application/json" \
    http://localhost:18888/api/v1beta1/proxies -d \
    '{ 
        "user": "user1", 
        "host": "host1", 
        "password": "123456", 
        "node": "node1", 
        "port": 2333 
    }'
    
    {
        "proxy_id": "8phwpv27",
        "user": "user1",
        "host": "host1",
        "password": "123456",
        "node": "node1",
        "port": 2333,
        "create_time": "2023-04-12T11:08:35Z",
        "update_time": "2023-04-12T11:08:35Z"
    }
```

## List reverse proxy rules
```shell
$ curl http://localhost:18888/api/v1beta1/proxies
# query optional: proxy_id
{
 "proxies": [
    {
      "proxy_id": "<proxy_id>",
      "user": "",
      "host": "host1:22",
      "password": "123456",
      "node": "j2001",
      "port": 18880,
      "create_time": "2023-04-12T09:35:39Z",
      "update_time": "2023-04-12T09:35:39Z"
    },
}

# Example
$ curl http://localhost:18888/api/v1beta1/proxies
{
  "proxies": [
    {
      "proxy_id": "localhost",
      "user": "user1",
      "host": "host1:22",
      "password": "123456",
      "node": "j2001",
      "port": 18880,
      "create_time": "2023-04-12T09:35:39Z",
      "update_time": "2023-04-12T09:35:39Z"
    },
    {
      "proxy_id": "127.0.0.1",
      "user": "user2",
      "host": "host2:22",
      "password": "123456",
      "node": "g0156",
      "port": 8888,
      "create_time": "2023-04-12T09:35:39Z",
      "update_time": "2023-04-12T09:35:39Z"
    }
  ]
}
```

## Delete a reverse proxy
```shell
$ curl -X DELETE http://localhost:18888/api/v1beta1/proxies/<proxy_id>

# Example
curl -X DELETE http://localhost:18888/api/v1beta1/proxies/localhost
```