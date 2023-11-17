# Proxy Log

`proxylog` is a tool to measure the the traffic of a webpage. It works as a socks4 proxy.

## Start

```bash
go build && ./proxylog
```

You have to configure the webbrowser to use a socks4 proxy on localhost:4567,
for example with https://addons.mozilla.org/de/firefox/addon/foxyproxy-standard/


Afterwars open http://localhost:9050 to see the sizes of content.
