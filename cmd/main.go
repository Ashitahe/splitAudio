package main

import (
	"bufio" // 用于高效的文件读取
	"bytes"
	"fmt" // 用于格式化输入输出
	"os"

	// 提供操作系统功能
	"os/exec" // 用于执行外部命令
	"path/filepath"
	"regexp" // 提供正则表达式支持
	"runtime"
	"strconv" // 用于字符串和基本数据类型之间的转换
	"strings"
	"sync"
)

var ffmpegPath string

func init() {
    ffmpegPath = findFFmpeg()
    if ffmpegPath == "" {
        fmt.Println("Error: ffmpeg not found. Please install ffmpeg and ensure it's in your PATH or in the current directory.")
        os.Exit(1)
    }
    fmt.Printf("Using ffmpeg from: %s\n", ffmpegPath)
}

func findFFmpeg() string {
    // 首先检查当前目录
    if _, err := os.Stat("ffmpeg"); err == nil {
        absPath, _ := filepath.Abs("ffmpeg")
        return absPath
    }
    if _, err := os.Stat("ffmpeg.exe"); err == nil {
        absPath, _ := filepath.Abs("ffmpeg.exe")
        return absPath
    }

    // 然后检查 PATH
    path, err := exec.LookPath("ffmpeg")
    if err == nil {
        return path
    }

    return ""
}

// detectSilence 函数执行静音检测并返回输出结果
func detectSilence(inputFile string) (string, error) {
	cmd := exec.Command(ffmpegPath, "-i", inputFile, "-af", "silencedetect=noise=-30dB:d=1", "-f", "null", "-")
	
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
			return "", fmt.Errorf("error running ffmpeg command: %v\nffmpeg output: %s", err, stderr.String())
	}
	
	return stderr.String(), nil
}

// parseSilence 函数从字符串解析静音时间段
func parseSilence(content string) [][2]float64 {
	var silenceEnds, silenceStarts []float64
	
	// 编译正则表达式用于匹配静音结束和开始时间
	reEnd := regexp.MustCompile(`silence_end: (\d+\.\d+)`)
	reStart := regexp.MustCompile(`silence_start: (\d+\.\d+)`)

	// 按行扫描内容
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
			line := scanner.Text()
			// 匹配静音结束时间
			if matches := reEnd.FindStringSubmatch(line); len(matches) > 1 {
					if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
							silenceEnds = append(silenceEnds, f)
					}
			}
			// 匹配静音开始时间
			if matches := reStart.FindStringSubmatch(line); len(matches) > 1 {
					if f, err := strconv.ParseFloat(matches[1], 64); err == nil {
							silenceStarts = append(silenceStarts, f)
					}
			}
	}

	min := func (a int, b int) int  {
		if (a < b) {
			return a
		}
		return b
	}

	var segments [][2]float64
	// 组合静音结束和下一个静音开始时间为一个段
	for i := 0; i < min(len(silenceEnds), len(silenceStarts)-1); i++ {
			segments = append(segments, [2]float64{silenceEnds[i], silenceStarts[i+1]})
	}

	return segments
}

// splitAudio 函数将输入的音频文件按照给定的时间段分割
func splitAudio(inputFile string, segments [][2]float64) error {
	dir := filepath.Dir(inputFile)
	filename := filepath.Base(inputFile)
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	for i, segment := range segments {
			start, end := segment[0], segment[1]
			outputFile := filepath.Join(dir, fmt.Sprintf("%s_part_%03d.mp3", nameWithoutExt, i))
			cmd := exec.Command(ffmpegPath, "-i", inputFile, "-ss", fmt.Sprintf("%f", start),
					"-to", fmt.Sprintf("%f", end), "-c", "copy", outputFile)
			err := cmd.Run()
			if err != nil {
					return fmt.Errorf("error processing segment %d of %s: %v", i, inputFile, err)
			}
	}
	return nil
}

func processMP3File(filePath string) error {
	fmt.Printf("Processing file: %s\n", filePath)
	silenceOutput, err := detectSilence(filePath)
	if err != nil {
			return fmt.Errorf("error detecting silence: %v", err)
	}

	segments := parseSilence(silenceOutput)
	if len(segments) == 0 {
			return fmt.Errorf("no silence detected in file: %s", filePath)
	}

	err = splitAudio(filePath, segments)
	if err != nil {
			return fmt.Errorf("error splitting audio: %v", err)
	}

	fmt.Printf("Successfully processed file: %s\n", filePath)
	return nil
}

func processDirectory(dirPath string, jobs chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
					return err
			}
			if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".mp3" {
					jobs <- path
			}
			return nil
	})

	if err != nil {
			fmt.Printf("Error walking the path %s: %v\n", dirPath, err)
	}
}

func worker(id int, jobs <-chan string, results chan<- error) {
	for filePath := range jobs {
			fmt.Printf("Worker %d processing file: %s\n", id, filePath)
			err := processMP3File(filePath)
			results <- err
	}
}

func main() {
	if len(os.Args) < 2 {
			fmt.Println("Usage: program <directory1> [directory2] [directory3] ...")
			os.Exit(1)
	}

	numWorkers := runtime.NumCPU() // 使用CPU核心数量作为工作线程数
	jobs := make(chan string, 100)
	results := make(chan error, 100)

	// 启动工作线程
	for w := 1; w <= numWorkers; w++ {
			go worker(w, jobs, results)
	}

	// 处理目录
	var wg sync.WaitGroup
	for _, dir := range os.Args[1:] {
			wg.Add(1)
			go processDirectory(dir, jobs, &wg)
	}

	// 等待所有目录处理完成
	go func() {
			wg.Wait()
			close(jobs)
	}()

	// 收集结果
	for i := 0; i < len(os.Args)-1; i++ {
			err := <-results
			if err != nil {
					fmt.Printf("Error: %v\n", err)
			}
	}
}