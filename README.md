# goradieschen

![Go Radieschen Logo](https://github.com/pilosus/goradieschen/blob/main/assets/logo_150px.png?raw=true)

Key-value store for Go with Redis-like API.

## Features

- Simple key-value store
- Key expiration
- Subset of Redis commands
- RESP2 protocol support

## Usage

1. Get the code & build the binary:

```shell
$ git clone https://github.com/pilosus/goradieschen.git && cd goradieschen && go build
```

2. Run the server:

```shell
$ ./goradieschen
```

3. Connect to the server using a Redis client:

```shell
$ redis-cli -h localhost -p 6380
```

4. Use the server with Redis-like commands:

```shell
localhost:6380> set k1 abc
OK
localhost:6380> get k1
"abc"
localhost:6380> expire k1 10
(integer) 1
localhost:6380> ttl k1
(integer) 7
```
