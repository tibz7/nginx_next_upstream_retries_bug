## Description
When Nginx is receiving enough traffic and an upstream server closes prematurely the connection,
the `proxy_next_upstream_tries` proxy directive is ignored. If enough requests causing some connections to close prematurely
are sent, all upstreams can potentially be unavailable. 
This seems to occur *only* when the connection is closed by the upstream server: for instance,if the application server returns a `500`
response and `proxy_next_upstream` is set to `http_500`, Nginx behaves correctly.  

#### How to reproduce the bug

# Start the stack
`docker compose up -d && docker compose ps`
Wait for the the build and start of the app servers

# Monitor nginx logs
`docker compose exec nginx-test tail -f /var/log/nginx/myapp.access.log /var/log/nginx/myapp.error.log`

# Send an OK request
`curl "htpp://localhost:2222/ok -v`
 Response:
 ``` 
    {"message":"ok"}% 
```
Nginx logs:
```
    ==> /var/log/nginx/myapp.access.log <==
    192.168.0.1 [01/Dec/2022:14:14:24 +0000] "GET /ok HTTP/1.1" 200 178 "-"  0.024
```

 # Send an kill request
 This request call the endpoint that closes the connection
`curl "htpp://localhost:2222/kill -v`
 
 Response:
 ``` 
    <html>
    <head><title>502 Bad Gateway</title></head>
    <body>
    <center><h1>502 Bad Gateway</h1></center>
    <hr><center>nginx</center>
    </body>
    </html>
 ```

 Nginx logs:
```
    ==> /var/log/nginx/myapp.access.log <==
    192.168.0.1 [01/Dec/2022:14:16:52 +0000] "GET /kill HTTP/1.1" 502 315 "-"  0.006

    ==> /var/log/nginx/myapp.error.log <==
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.3:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.4:8080/kill", host: "localhost:2222"
```

As expected nginx follow the directive `proxy_next_upstream_tries 2;` and retries two upstreams servers.
If you make this call repeatedly, you will keep seeing the exact same message, which is the intended behavour

## Load Nginx
Generate some traffic on Nginx. The script is only calling the good endpoint.
`docker run -v "$(pwd)"/load_test.js:/tmp/load_test.js  grafana/k6 run /tmp/load_test.js`

You should see a lot of `200` requests in the `access.log` but no errors in the `error.log`

## Send an other kill request
`curl "htpp://localhost:2222/kill -v`
You should see 6 errors (one per upstream server), the `proxy_next_upstream_tries` is ignored:
```
    ==> /var/log/nginx/myapp.access.log <==
    192.168.0.1 [01/Dec/2022:14:16:52 +0000] "GET /kill HTTP/1.1" 502 315 "-"  0.006

    ==> /var/log/nginx/myapp.error.log <==
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.3:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.4:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.5:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.6:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.7:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.8:8080/kill", host: "localhost:2222"
```

If you repeat this command while nginx is under load, nginx will keep on retrying 6 times. 
If too many requests causing a `connection closed` are sent, you will eventually run out of available upstreams for a period of time defined by `fail_timeout=10s`:
```
    2022/12/01 15:12:19 [error] 237#237: *207 no live upstreams while connecting to upstream, client: 192.168.112.1, server: myapp.com, request: "GET /ok HTTP/1.1", upstream: "http://backend_servers/ok", host: "host.docker.internal:2222"
```
The `no live upstream` is of course an expectable consequence.
At this point you can stop `k6` and wait for an other 60s before calling again `/kill`. There will be only two error messages: the `proxy_next_upstream_tries` is again considered.

 Nginx logs:
```
    ==> /var/log/nginx/myapp.access.log <==
    192.168.0.1 [01/Dec/2022:14:16:52 +0000] "GET /kill HTTP/1.1" 502 315 "-"  0.006

    ==> /var/log/nginx/myapp.error.log <==
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.3:8080/kill", host: "localhost:2222"
    2022/12/01 14:16:52 [error] 113#113: *17 upstream prematurely closed connection while reading response header from upstream, client: 192.168.0.1, server: myapp.com, request: "GET /kill HTTP/1.1", upstream: "http://192.168.0.4:8080/kill", host: "localhost:2222"
```

