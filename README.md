# gohosts
A simple windows tool to switch hosts quickly!

![screenshot: `gohosts `](preview.png)


# How to build in windows

1. download [sciter's SDK](http://sciter.com/download/)  

2. put ```bin\sciter32.dll``` and ```bin\sciter64.dll``` in windows system32 directory

3. install [MSYS2](https://msys2.github.io/), open a MSYS2 shell and execute:  

```
windres -o rsrc.syso resource.rc
go run cmd/dist.go
go build -ldflags -H=windowsgui
```