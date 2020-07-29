# BOS CMD

百度智能云已经提供了基于BCE Python SDK的命令行工具 [BCE CLI](https://cloud.baidu.com/doc/BOS/s/zjwvyqm6d)（Command Line Interface），但是Python 版BCE CLI 依赖用户的操
作环境，而且需要用户安装Python 2.7，所以百度智能云现在向用户提供一种安装更方便，执行效率更高，而且使用方法与 BCE CLI 相
同的命令行工具 —— BCE CMD。

BCE CMD当前只包含了BOS的功能，所以下文都用BOS CMD代指BCE CMD。BOS CMD提供了丰富的功能，您不仅可以通过BOS CMD完成Bucket的
创建和删除，Object的上传、下载、删除和拷贝等, 当您从BOS下载文件或者上传文件到BOS遇到问题时，还可以使用BCE CMD的子命令
bosprobe进行错误检查。

## 快速开始

### 编译

编译前提：

1. 安装 golang (至少为 1.11);
2. 配置环境变量 GOROOT；

执行下面的代码，快速编译：

```
sh build.sh
```

编译产出在目录 ./output 中.

### 配置和使用 bcecmd

请参考 [BCECMD官方文档](https://cloud.baidu.com/doc/BOS/s/Sjwvyqetg)

## 测试

```
go test -v $1 -coverprofile=c.out #请将 $1 替换你需要测试的文件或目录
go tool cover -html=c.out -o cover.html
```

## 如何贡献

迎您修改和完善此项目，请直接提交PR 或 issues。

* 提交代码时请保证良好的代码风格。
* 提交 issues 时， 请翻看历史 issues， 尽量不要提交重复的issues。

## 讨论

欢迎提 issues。
