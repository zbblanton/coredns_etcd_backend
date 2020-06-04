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
# Notes

* Still working on wildcard support
* SOA will return a static record
* No reverse lookup at this time