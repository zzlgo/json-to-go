# 私有化部署

建议优先使用在线地址，json2go地址（chrome）: https://zzlgo.github.io/json2go

## 编译

```text
go版本1.19.3
tinygo版本0.27.0

# 编译产生json2go.tar.gz
make build
```

## 部署

采用docker nginx部署，nginx的default.conf配置如下

```text
server {
    listen       80;
    listen  [::]:80;
    server_name  localhost;


    location / {
        root   /usr/share/nginx/html;
        index  index.html index.htm;
    }

    location /json2go {
    alias /usr/share/nginx/html/json2go/;
    index index.html;
        try_files $uri $uri/index.html /json2go/index.html;
    }

    #error_page  404              /404.html;

    # redirect server error pages to the static page /50x.html
    #
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }

}
```

启动docker，将default.conf和json2go文件夹挂载到nginx容器，访问地址为http://localhost:8080/json2go

```text
docker run --restart=on-failure -d -p 8080:80 -p 443:443 -v /data/nginx/default.conf:/etc/nginx/conf.d/default.conf -v /data/nginx/json2go:/usr/share/nginx/html/json2go --name nginx_json2go nginx:1.22
```
