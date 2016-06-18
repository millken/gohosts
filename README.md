# gohosts
Change windows hosts quickly as thunber!


# How to build in windows

```
// Need installed MSYS2, open a MSYS2 shell and execute:
windres -o rsrc.syso resource.rc
go build -ldflags -H=windowsgui
```