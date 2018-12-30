# yuque-sync
[![Build Status](https://travis-ci.org/gudegg/yuque-sync.svg?branch=user_conf)](https://travis-ci.org/gudegg/yuque-sync)


同步[语雀](https://www.yuque.com)markdown格式到本地,支持hexo、hugo静态博客格式,只支持公开下载


# 使用

```shell
yuque-sync -n=namespace [options:-o、-p、-t]
```
namespace说明:
如`https://www.yuque.com/yuque/help`,则namespace为`yuque/help`,详细查看官方文档[Repo - 仓库](https://www.yuque.com/yuque/developer/repo),可以从
[https://www.yuque.com/api/v2/users/yuque/repos](https://www.yuque.com/api/v2/users/yuque/repos)获取所有你公开的namespace,链接中的`yuque`替换成你自己的登录名
```
Usage of yuque-sync:
  -n   string
          设置namespace
  [-o] bool
          文件同名覆盖写入 (default true)
  [-p] string
          设置存储路径 (default "download/")
  [-t] string
          文件写入格式,支持hexo、hugo (default "raw")
```