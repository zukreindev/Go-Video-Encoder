package main

import (
	"archive/zip"
	"fmt"
	"github.com/fatih/color"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	color.Blue("GO Video Encoder")
	isFFMPEGInstalled := isFFMPEGInstalled()

	if !isFFMPEGInstalled {
		install := installFFMPEG()
		if install {
			color.Green("ffmpeg installed")
			main()
		} else {
			color.Red("ffmpeg not installed")
		}
	}

	videoInfo := getVideoInfo()

	if videoInfo != (Video{}) {
		encode := EncodeVideo(videoInfo.Name, videoInfo.Codec, videoInfo.CompressionLevel)

		if encode {
			color.Green("Video encoded")
			fmt.Println("Press enter to exit")
			fmt.Scanln()
		} else {
			color.Red("Video not encoded")
			fmt.Println("Press enter to exit")
			fmt.Scanln()
		}
	}

}

type Video struct {
	Name             string
	Codec            string
	CompressionLevel string
}

func getVideoInfo() Video {
	color.HiCyan("Copy your video to the videos folder")

	videosDir := filepath.Join("videos")

	if _, err := os.Stat(videosDir); os.IsNotExist(err) {
		os.Mkdir(videosDir, 0755)
	}

	color.Green("Enter a video name:")

	var videoName string

	fmt.Scanln(&videoName)

	if string(videoName[len(videoName)-4:]) != ".mp4" {
		color.Red("Invalid video format")
		getVideoInfo()
	}

	color.Green("Enter a codec name (h264, libx265, libvpx, libvpx-vp9):")

	var codecName string

	fmt.Scanln(&codecName)

	if codecName != "h264" && codecName != "libx265" && codecName != "libvpx" && codecName != "libvpx-vp9" {
		color.Red("Invalid codec name")
		getVideoInfo()
	}

	color.Green("Enter a compression level (low, medium, high):")

	var compressionLevel string

	fmt.Scanln(&compressionLevel)

	if compressionLevel == "low" {
		compressionLevel = "23"
	} else if compressionLevel == "medium" {
		compressionLevel = "18"
	} else if compressionLevel == "high" {
		compressionLevel = "10"
	} else {
		color.Red("Invalid compression level")
		getVideoInfo()
	}

	return Video{
		Name:             videoName,
		Codec:            codecName,
		CompressionLevel: compressionLevel,
	}
}

func EncodeVideo(videoName string, codecName string, compressionLevel string) bool {
	video := filepath.Join("videos", videoName)

	if _, err := os.Stat(video); os.IsNotExist(err) {
		color.Red("Video not found in videos folder")
		return false
	}

	ffmpegDir := filepath.Join("bin", "ffmpeg.exe")
	outputName := "videos/"+videoName[:len(videoName)-4] + "_encoded.mp4"

	if _, err := os.Stat(outputName); !os.IsNotExist(err) {
		os.Remove(outputName)
	}

	cmd := exec.Command(ffmpegDir, "-i", video, "-c:v", codecName, "-crf", compressionLevel, "-c:a", "copy", outputName)

	err := cmd.Run()

	if err != nil {
		panic(err)
	}

	return true
}

func isFFMPEGInstalled() bool {
	ffmpegDir := filepath.Join("bin", "ffmpeg.exe")
	ffprobeDir := filepath.Join("bin", "ffprobe.exe")

	if _, ffmpegError := os.Stat(ffmpegDir); os.IsNotExist(ffmpegError) {
		color.Red("ffmpeg.exe not found in bin folder")
		return false
	}

	if _, ffprobeError := os.Stat(ffprobeDir); os.IsNotExist(ffprobeError) {
		color.Red("ffprobe.exe not found in bin folder")
		return false
	}

	return true
}

func installFFMPEG() bool {

	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		os.Mkdir("temp", 0755)
	}

	color.Yellow("Installing ffmpeg")

	out, err := os.Create("temp/ffmpeg.zip")

	if err != nil {
		panic(err)
	}

	defer out.Close()

	// Get the FFMPEG zip file
	resp, err := http.Get("https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip")

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)

	if err != nil {
		panic(err)
	}

	color.Green("ffmpeg installed")

	color.Yellow("Unzipping ffmpeg")

	err = extractBinFolder("temp/ffmpeg.zip")

	if err != nil {
		panic(err)
	}

	color.Green("ffmpeg unzipped")

	os.Remove("temp/ffmpeg.zip")

	return true
}

func extractBinFolder(zipFilePath string) error {
	r, err := zip.OpenReader(zipFilePath)
	if err != nil {
		return err
	}
	defer r.Close()

	err = os.MkdirAll("bin", os.ModePerm)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		if f.Name == "ffmpeg-master-latest-win64-gpl/bin/ffmpeg.exe" || f.Name == "ffmpeg-master-latest-win64-gpl/bin/ffprobe.exe" || f.Name == "ffmpeg-master-latest-win64-gpl/bin/ffplay.exe" {
			rc, err := f.Open()

			if err != nil {
				return err
			}
			fmt.Println(f.Name)

			defer rc.Close()

			path := filepath.Join("bin", f.Name)

			switch f.Name {
			case "ffmpeg-master-latest-win64-gpl/bin/ffmpeg.exe":
				path = filepath.Join("bin", "ffmpeg.exe")
			case "ffmpeg-master-latest-win64-gpl/bin/ffprobe.exe":
				path = filepath.Join("bin", "ffprobe.exe")
			case "ffmpeg-master-latest-win64-gpl/bin/ffplay.exe":
				path = filepath.Join("bin", "ffplay.exe")
			}

			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())

			if err != nil {
				return err
			}

			defer f.Close()

			_, err = io.Copy(f, rc)

			if err != nil {
				return err
			}

			f.Close()
		}
	}

	err = r.Close()
	if err != nil {
		return err
	}

	return nil
}
