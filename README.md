# Video to ASCII in Go

    Converts video to ASCII
![](img/img1.png)

### Build

```shell
go mod tidy
go build .
```

### Run

> Requires ffmpeg binary in $PATH (or -ffmpeg path/to/bin/ffmpeg)

```shell
go mod tidy
go run .
```

### Launch parameters

- `-video path/to/video` -> specify video
- `-ffmpeg path/to/bin/ffmpeg` -> specify ffmpeg binary