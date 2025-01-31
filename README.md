# Video to ASCII in Go

    Converts video to ASCII
![](img/img1.png)

### Build && Run (Windows)

> Requires ffmpeg binary in $PATH

```shell
go mod tidy
go run -ldflags "-s -w" cmd/app/main.go cmd/app/init_windows.go
```

### Launch parameters

- `-video path/to/video` -> specify video
- `-ffmpeg path/to/bin/ffmpeg` -> specify ffmpeg binary