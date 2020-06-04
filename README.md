# CoreDNS Etcd Backend

This is another etcd backend for coredns. I built this plugin for coredns becuase the current etcd plugin does not support multiple subdomains (https://github.com/coredns/coredns/issues/2941). This plugin is based off the current Etcd plugin and the external redis plugin.

# Using this plugin

To use this plugin you need to recompile CoreDNS with this plugin enabled. This is really easy with the use of go modules.

Clone the CoreDNS repo:

``` sh
git clone https://github.com/coredns/coredns
cd coredns
```

Modify the plugins.cfg and add this plugin to it. Something like this:

```
...
etcd:etcd
coredns_etcd_backend:github.com/zbblanton/coredns_etcd_backend
redis:github.com/arvancloud/redis
sign:sign
...
```

Now just run:

```sh
make
```

Configuring the plugin works the exact same way the etcd plugin does. Look in the `Syntax` section of https://coredns.io/plugins/etcd/

Example:
```
awesomedomain.com:4000 {
    log
    errors
    coredns_etcd_backend {
        endpoint  http://127.0.0.1:2379
    }
}
```

Example with TLS:
```
awesomedomain.com:4000 {
    log
    errors
    etcd {
        endpoint  https://10.0.0.10:2379
        tls /data/coredns-etcd.crt /data/coredns-etcd.key /data/ca.crt
    }
}
```

# Records

Create records by the reverse name. For example if you want to create an `A` record for test1.test.local you would do a put to etcd for `/coredns/local/test/test1/-A-/record1`. There are some exceptions to the last path name. For example for `CNAME` records it will end like `/coredns/local/test/test1/-CNAME-`.

NOTE: the name for record doesn't matter. The `record1` above could be any name.

## A

``` json
{
  "host": "10.0.0.102"
}
```

Example:

```
/coredns/local/test/test1/-A-/record1
{"host":"10.0.0.102"}
/coredns/local/test/test1/-A-/record2
{"host":"10.0.0.103"}
```

## CNAME

``` json
{
  "cname": "testme.local"
}
```

Example:

```
/coredns/local/test/test1/-CNAME-
{"cname":"testme.local"}
```

## TXT

``` json
{
  "text": "im a test 2"
}
```

Example:

```
/coredns/local/test/test1/-TXT-/record1
{"text":"im a test"}
/coredns/local/test/test1/-TXT-/record2
{"text":"im a test 2"}
```

## SRV

``` json
{
  "target": "test.local",
  "weight": 10,
  "port": 2222,
  "priority": 0
}
```

Example:

```
/coredns/local/test/test1/test2/_tcp/_etcd-server-ssl/-SRV-/record1
{"target":"test.local","weight":10,"port":2222,"priority":0}
/coredns/local/test/test1/test2/_tcp/_etcd-server-ssl/-SRV-/record2
{"target":"test2.local","weight":10,"port":2222,"priority":10}
```

# Notes

* Still working on wildcard support
* SOA will return a static record
* No reverse lookup at this time