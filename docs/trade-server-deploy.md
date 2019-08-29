# trade-server部署说明

## 1. kafka部署

kafka依赖zookeeper组件。可以通过源码、docker等方式安装部署。

源码安装参考：https://kafka.apache.org/quickstart

docker安装参考：https://hub.docker.com/r/wurstmeister/kafka/

mac可通过brew安装参考 https://github.com/coinexchain/trade-server/blob/master/examples/README.md

## 2. dex-node部署及配置

节点部署参考 https://github.com/coinexchain/testnets/blob/master/coinexdex-test/testnet-guide.md

部署完成后，需要修改配置文件``app.toml`` (默认在~/.cetd/config/app.toml)，添加以下配置

```toml
feature-toggle = true
subscribe-modules = "comment,authx,bankx,market,bancorlite"
brokers = "kafka:127.0.0.1:9092"
```

brokers配置按实际kafka部署配置填写。修改完后重启节点。

## 3. trade-server部署

### 编译

需要安装Go (推荐12及以上版本) https://golang.org/doc/

进入工程目录，执行以下命令进行编译

```shell
GO111MODULE=on go install ./...
```

编译完成后，二进制文件在 $GOPATH/bin 下

### 配置文件说明

工程根目录提供了默认配置模板文件`config.toml.default`，可在此基础上拷贝修改

```toml
# 监听端口
port = 8000

# 是否代理DEX节点的REST API
proxy = false 

# DEX节点LCD地址
lcd = "http://localhost:1317"

# Kafka地址
kafka-addrs = "localhost:9092"

# LevelDB数据目录
data-dir = "data"

# Log文件目录
log-dir = "log"

# Log级别: debug | info | warn | error
log-level = "info"

# Log格式: plain(普通文本格式) | json (json格式)
log-format = "plain"
```

启动时通过-c参数指定配置文件路径

```bash
nohup trade-server -c config.toml &
```