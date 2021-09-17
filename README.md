## log agent

本地日志备份到阿里云oss

- 支持加密备份文件
- 支持cron表达式定时上传

配置文件说明

```text
{
  "hostname": "", // 主机名, 默认为机器ip地址
  "endpoint": "http://oss-cn-shenzhen.aliyuncs.com", // oss yourEndpoint
  "accessKey": "xxx", // oss yourAccessKeyId
  "secretKey": "xxx", // oss yourAccessKeySecret
  "bucketName": "app-logstore", // oss yourBucketName
  "partSize": 102400, // 分片大小
  "log": {
    "path": "/tmp/oss/logs/",
    "rotate": 3
  },
  "strategies": [ // 备份策略
    {
      "src": "/tmp/logs/user/", // 源路径
      "dest": "/sunline/user/${hostname}/", // oss存储路径
      "pattern": "", // 匹配策略
      "exclude": "", // 排除策略
      "forbidWrite": true, // oss对象已经存在时, 是否允许覆盖
      "afterDelete": true, // 备份完成之后, 是否删除原文件
      "beforeTime": 3000, // 只上传修改时间在3000秒之前的文件
      "sign": "123456", // zip 加密字符串
      "spec": "* */5 * * * *" // cron表达式
    }
  ]
}

```

使用systemd管理服务

```text
将压缩包解压到 /usr/local/log-agent/ 目录, 创建配置文件 /usr/local/log-agent/config/app.json，然后执行以下命令

cd /usr/local/log-agent/
cp docs/systemd/log-agent.service /usr/lib/systemd/system/log-agent.service
systemctl daemon-reload
systemctl start log-agent
systemctl enable log-agent
systemctl status log-agent
```
