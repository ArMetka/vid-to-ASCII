package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sqweek/dialog"
	"golang.org/x/term"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const brightnessStr = "  `.-':_,^=;><+!rc*/z?sLTv)J7(|Fi{C}fI31tlu[neoZ5Yxjya]2ESwqkP6h9d4VpOGbUAKXHm8RD#$Bg0MNWQ%&@"
const fps = 24

var ffmpeg string
var pathToFile string
var err error = nil

/*
	 ./program
		-ffmpeg path_to_ffmpeg
		-video path_to_video
*/
func main() {
	fmt.Print("\033[H\033[2J") // Clear screen move to 0,0
	fmt.Print("\033[s")        // Save cursor position

	fmt.Println("Select a video file")

	// Parse args
	parseArgs()
	if pathToFile == "" {
		pathToFile, err = dialog.File().Load()
		if err != nil {
			panic("Error while selecting file!\n")
		}
	}
	//fmt.Println(ffmpeg)
	//fmt.Println(pathToFile)

	// Check if ffmpeg installed
	if ffmpeg == "" {
		_, err = exec.LookPath("ffmpeg")
		ffmpeg = "ffmpeg"
		if err != nil {
			panic("ffmpeg not found in $PATH: https://www.ffmpeg.org/download.html\n")
		}
	}

	// Check if in terminal
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		panic("Not a terminal!\n")
	}

	// Get and print terminal dimensions
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic("Error while getting size of the terminal!\n")
	}
	height -= 1 // extra info line
	fmt.Printf("Current terminal dimensions: %d x %d\n", width, height)

	// Get path to a video file
	if pathToFile == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("Enter path to video file: ")
		if scanner.Scan() {
			pathToFile = filepath.FromSlash(scanner.Text())
		}

		// Check if file exist
		if _, err = os.Stat(pathToFile); errors.Is(err, os.ErrNotExist) {
			panic("File does not exist!\n")
		}
	}

	// Convert video to image sequence
	var frameCount int
	err = convertMp4ToImgSeq(&frameCount)
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 2)

	initScreen(width, height)
	// Read image, convert, print
	deltaTime := time.Duration(1000/fps) * time.Millisecond // frame time in ms
	for i := 0; i < frameCount; i++ {
		wakeUpTime := time.Now().Add(deltaTime)

		// Read
		imgPath := filepath.Join("imgs", "out"+strconv.Itoa(i+1)+".png")
		var img image.Image
		readImage(&img, imgPath)

		// Convert and print
		convertAndPrint(&img, width, height)

		// Print info in the bottom
		printInfo(i, fps, width, height)

		// Resize every 0.5 sec
		if i%(fps/2) == 0 {
			width, height, _ = term.GetSize(int(os.Stdout.Fd()))
			height -= 1 // extra info line
		}

		time.Sleep(time.Until(wakeUpTime))
	}

	time.Sleep(time.Second * 3)
}

func convertMp4ToImgSeq(frameCount *int) error {
	// Create dir (or check if exist)
	err = os.Mkdir("imgs", os.ModePerm)
	if (err != nil) && (!errors.Is(err, os.ErrExist)) {
		panic("Failed to create imgs directory!\n")
	}
	outFiles := filepath.Join("imgs", "out%d.png")

	fmt.Print("Processing...")
	timeStart := time.Now()

	// ffmpeg -i video.mp4 out%d.png
	out, err := exec.Command(ffmpeg, "-i", pathToFile, "-vf", "fps="+strconv.Itoa(fps), outFiles).CombinedOutput()
	if err != nil {
		fmt.Print(string(out))
	} else {
		fmt.Print("\nDone.\n")

		// Get number of frames
		frameIndex := strings.LastIndex(string(out), "frame= ")
		fpsIndex := strings.LastIndex(string(out), "fps=")
		*frameCount, err = strconv.Atoi(strings.TrimSpace(string(out[frameIndex+6 : fpsIndex])))
		if err != nil {
			panic("Failed to extract frame count!\n")
		}
		fmt.Printf("ffmpeg: frames = %d\n", *frameCount)

		// Print conversion time
		fmt.Printf("ffmpeg: elapsed time = %.0fs\n", time.Since(timeStart).Seconds())
	}

	return err
}

func readImage(img *image.Image, imgPath string) {
	// Open file
	imgFile, err := os.Open(imgPath)
	if err != nil {
		panic("Failed to open `" + imgPath + "`!\n")
	}
	defer imgFile.Close()

	// Read png
	*img, err = png.Decode(imgFile)
	if err != nil {
		panic("Failed to decode png file `" + imgPath + "`!\n")
	}
}

func initScreen(width int, height int) {
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			out.WriteString(" ")
		}
		out.WriteString("\n")
	}
}

func convertAndPrint(img *image.Image, width int, height int) {
	fmt.Print("\033[u") // Restore cursor position

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	bounds := (*img).Bounds()
	wRatio, hRatio := float32(bounds.Max.X)/float32(width), float32(bounds.Max.Y)/float32(height)
	var w, h int // Current pixel on actual image

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			w, h = int(float32(j)*wRatio), int(float32(i)*hRatio)
			r, g, b, _ := (*img).At(w, h).RGBA()       // rgba values are [0, 0xffff]
			brightness := int((r + g + b) / (3 * 257)) // convert to brightness [0, 0xff]
			out.WriteString(convertToAscii(brightness))
		}
		out.WriteString("\n")
	}
}

func convertToAscii(brightness int) string {
	index := 0
	if brightness >= 255 {
		index = len(brightnessStr) - 1
	} else if brightness > 0 {
		index = int(float32(len(brightnessStr)) / 255.0 * float32(brightness))
	}
	return string(brightnessStr[index])
}

func printInfo(frame int, fps int, width int, height int) {
	sb := strings.Builder{}
	sb.WriteString("frame=")
	sb.WriteString(strconv.Itoa(frame))
	sb.WriteString(" fps=")
	sb.WriteString(strconv.Itoa(fps))
	sb.WriteString(" out=")
	sb.WriteString(strconv.Itoa(width))
	sb.WriteString("x")
	sb.WriteString(strconv.Itoa(height))
	toFill := width - sb.Len()
	if toFill < 0 {
		return
	}
	for i := 0; i < toFill; i++ {
		sb.WriteString(" ")
	}
	s := sb.String()
	fmt.Print(s)
}

func parseArgs() {
	switch len(os.Args) {
	case 1:
		return
	case 3, 5:
		for i := 1; i < len(os.Args); i += 2 {
			if os.Args[i] == "-ffmpeg" {
				_, err := exec.LookPath(os.Args[i+1])
				if err != nil {
					panic("Invalid ffmpeg path!\n")
				}
				ffmpeg = os.Args[i+1]
			}
			if os.Args[i] == "-video" {
				if _, err := os.Stat(os.Args[i+1]); errors.Is(err, os.ErrNotExist) {
					panic("Invalid video path!\n")
				}
				pathToFile = os.Args[i+1]
			}
		}
	default:
		panic("Invalid program arguments!\n")
	}
}
