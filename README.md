# B2WebP

极简 WebP 图床，使用Golang+SQLite，可以单文件部署无需外部数据库。

依赖`cwebp`命令行程序实现服务器端图片转换，支持客户端转换webp图片后上传。

该程序的主要数据会存储在S3中，频繁访问的热点数据会在本地缓存一份提供服务，尽可能的防止意外的S3账单发生。

已在Backblaze B2上得到良好测试。

## 配置项
该程序使用环境变量控制行为

| 变量名 | 默认值 | 含义 |
|--|--|--|
|B2WEBP_SITE_NAME | B2WEBP | 站点标题 |
|B2WEBP_DEBUG |false | 调试模式 |
|B2WEBP_LISTEN | 127.0.0.1:9000 | 侦听地址 |
|B2WEBP_DATA_PATH| ./data |数据存储地址|
|B2WEBP_HASHID_SALT|B2WEBP|哈希ID使用的盐值|
|B2WEBP_S3_MAX_CACHE_SIZE|500MB| 本地S3缓存容量上限 |
|B2WEBP_MEMORY_CACHE_SIZE|50MB| 信息缓存容量上限|
|B2WEBP_ASSETS_PATH|assets|资源文件地址|
|B2WEBP_CDN_ASSETS_PREFIX|https://cdn.jsdelivr.net/npm|前端资源CDN前缀|
|B2WEBP_BASE_URL||基础地址，配置为访问域名即可|
|B2WEBP_MOTD||今日提示，显示在上传页面顶部|
|B2WEBP_UPLOAD_PASSWORD||上传密码，设置后会要求密码|
|B2WEBP_S3_REGION||S3存储桶区域|
|B2WEBP_S3_ENDPOINT||S3远程端点|
|B2WEBP_S3_BUCKET||S3存储桶名|
|B2WEBP_S3_ACCESS_ID||S3访问ID|
|B2WEBP_S3_ACCESS_KEY||S3访问密钥|
|B2WEBP_S3_OBJECT_PREFIX||S3存储对象前缀|

支持读取当前目录下的.env文件加载环境变量进行持久配置
