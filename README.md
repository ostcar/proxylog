# Proxy Log

`proxylog` is a tool to measure the the traffic of a webpage. It works as a socks4 proxy.

## Start

```bash
go build && ./proxylog
```

You have to configure the webbrowser to use a socks4 proxy on localhost:4567,
for example with https://addons.mozilla.org/de/firefox/addon/foxyproxy-standard/


Afterwars open http://localhost:9050 to see the sizes of content.


## Log sizes to file

To log all traffic size so a file, call the tool with a filename. It will be
opened in append only, so it is save to open an existing file:


```bash
go build && ./proxylog sizes.log
```
