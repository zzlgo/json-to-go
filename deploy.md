# 私有化部署

建议优先使用在线地址，json-to-go地址（chrome）: https://zzlgo.github.io/json-to-go

## 编译

```text
go版本1.19.3
tinygo版本0.27.0
make build
```

## 部署

```text
make docker
```
访问地址为http://localhost:8080/json-to-go

```text
docker run --restart=on-failure -d -p 8080:80 -p 443:443  --name json-to-go json-to-go
```


