package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"os"
	"regexp"
	"time"
)

const (
	// Success 成功
	Success = iota

	// PathNullErr 路径为空
	PathNullErr    

	// InvalidPathErr 不是有效的路径
	InvalidPathErr

	// FileNotDIR 是文件，而不是目录
	FileNotDIR

	// OutputPathErr 输出文件的路径错误
	OutputPathErr

	// PermissionErr 权限错误
	PermissionErr

	// NoChildrenErr 即没有子目录，也没有文件
	NoChildrenErr
)

var hashChanel = make(chan *Node, 100)

// MaxWriterCount 同时写hash的最大并发数
const MaxWriterCount = 100

// OutputPath 默认的输出文件路径
const OutputPath = "treehash.txt"

func createWriter(output string) {
	for node := range hashChanel {
		file, err := os.Open(node.Path)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(-1)
		}
		hash := sha1.New()
		if _, ioErr := io.Copy(hash, file); ioErr != nil {
			fmt.Println(ioErr.Error())
			os.Exit(-1)
		}

		if closeErr := file.Close(); closeErr != nil {
			fmt.Println(closeErr.Error())
			os.Exit(-1)	
		}
		
		fd, fileErr := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if fileErr != nil {
			fmt.Println("hash file error: ", fileErr.Error())
			os.Exit(-1)
		}

		data := fmt.Sprintf("%s,%x,%d\n", node.Path, hash.Sum(nil), node.Size)
		buf  := []byte(data)
		fd.Write(buf)
		if fdErr := fd.Close(); fdErr != nil {
			fmt.Println(fdErr.Error())
			os.Exit(-1)	
		}
	}
}

// Node 树中的结点
type Node struct {
	ParentDir  string //父目录的路径
	Path       string //当前结点的完整路径
	Name       string //当前结点的文件名
	Size       int64  //文件大小
	IsDir      bool   //是否是目录
}

func newNode(fileInfo os.FileInfo, parentDir string) Node {
	node := Node{
		ParentDir : parentDir,
		Path      : parentDir + string(os.PathSeparator) + fileInfo.Name(),
		Name      : fileInfo.Name(),
		Size      : fileInfo.Size(),
		IsDir     : fileInfo.IsDir(),
	}
	return node
}

// Traverse 遍历目录
func Traverse(path string, filter string, output string) int {
	if path == "" {
		fmt.Println("root参数不能为空")
		displayHelpCMD()
		return PathNullErr	
	}
	rootDir, err := os.Stat(path)
	if err != nil {
		fmt.Println(path + " 不是有效的目录")
		displayHelpCMD()
		return InvalidPathErr
	}
	if !rootDir.IsDir() {
		fmt.Println(path + " 不是目录")
		displayHelpCMD()
		return FileNotDIR
	}

	if (output != "") {
		outputFile, outErr := os.OpenFile(output, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if outErr != nil {
			fmt.Println("output error: ", outErr.Error())
			displayHelpCMD()
			return OutputPathErr
		}
		if closeErr := outputFile.Close(); closeErr != nil {
			fmt.Println("output error: ", closeErr.Error())
			displayHelpCMD()
			return OutputPathErr
		}
	} else {
		output = OutputPath
	}

	// 遍历时，保存树中的结点
	var stack []Node
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Println(err.Error())
		return PermissionErr
	}
	length := len(files)
	if length == 0 {
		fmt.Println(path + " 目录下即没有子目录，也没有文件")
		displayHelpCMD()
		return NoChildrenErr
	}

	var reg *regexp.Regexp
	var regErr error

	if filter != "" {
		if reg, regErr = regexp.Compile(filter); regErr != nil {
			reg = nil
		}
	}

	for i := 0; i < length; i++ {
		stack = append(stack, newNode(files[i], path))
	}

	for i := 0; i < MaxWriterCount; i++ {
		go createWriter(output)
	}

	// 树的广度优先遍历
	for len(stack) > 0 {
		node := stack[0]
		stack = stack[1:]
		if reg != nil && reg.MatchString(node.Name) {
			continue
		}
		if !node.IsDir {
			hashChanel <- &node
		} else {
			files, err := ioutil.ReadDir(node.Path)
			if err != nil {
				fmt.Println(err.Error())
				return PermissionErr
			}
			if len(files) > 0 {
				for i := 0; i < len(files); i++ {
					stack = append(stack, newNode(files[i], node.Path))
				}
			}
		}
	}
	close(hashChanel)
	return Success
}

func displayHelpCMD() {
	fmt.Println("运行以下命令获得帮助")
	fmt.Println("go run main.go help")
}

func displayHelp() {
	fmt.Println("*********************************************")
	fmt.Println("*  参数说明:                                *")
	fmt.Println("*  -root", "要计算hash的根目录                 *")
	fmt.Println("*  -filter", "需要过滤的目录或文件，支持通配符 *")
	fmt.Println("*  -output", "最后写入的文件路径               *")
	fmt.Println("*********************************************")
}

func main() {
	beginTime := time.Now()
	root      := flag.String("root", "", "要生成hash树的根目录")
	filter    := flag.String("filter", "", "过滤目录或文件，支持通配符")
	output    := flag.String("output", "", "最后写入的文件路径")

	flag.Parse()

	args    := flag.Args()
	hasHelp := false
	if len(args) >= 1 {
		for _, value := range args {
			if value == "help" {
				hasHelp = true
				displayHelp()
				break;
			}
		}
	}

	if hasHelp && *root == "" {
		os.Exit(0)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	if result := Traverse(*root, *filter, *output); result != Success {
		os.Exit(-1)
	}
	fmt.Println("duration: ", time.Now().Sub(beginTime).Seconds(), "s")
}
