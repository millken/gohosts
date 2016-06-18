# gohosts
switch hosts quickly as lightning!

![screenshot: `gohosts `](preview.png)


# How to build in windows

1. download [sciter's SDK](http://sciter.com/download/)  

2. install [MSYS2](https://msys2.github.io/), open a MSYS2 shell and execute:  

```
windres -o rsrc.syso resource.rc
go build -ldflags -H=windowsgui
```