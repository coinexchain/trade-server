# https configuration
trade-server默认不开启https，要开启的话需要修改配置文件，例如默认配置文件config.toml

```
# https
https-toggle = false
cert-dir = "cert"
```

修改https-toggle为true，修改cert-dir为服务器证书所在目录，或将证书和私钥复制到该配置目录下。

如果证书是本地自签名证书，使用curl访问server时，添加-k参数跳过server证书验证。

### FAQ

Q：请求server时报类似curl: (35) error:140770FC:SSL routines:SSL23_GET_SERVER_HELLO的错误

A：检查配置文件是否打开https开关



