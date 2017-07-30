package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"
	"os"
	"sync"
	"time"
)

var mu sync.Mutex

// Node 树中的结点
type Node struct {
	ParentDir  string //父目录的路径
	Path       string //当前结点的完整路径
	Name       string //当前结点的文件名
	Size       int64  //文件大小
	IsDir      bool   //是否是目录
	Hash       string
}

const (
	// Success 成功
	Success = iota

	// PathNullErr 路径为空
	PathNullErr    

	// InvalidPathErr 不是有效的路径
	InvalidPathErr

	// FileNotDIR 是文件，而不是目录
	FileNotDIR

	// PermissionErr 权限错误
	PermissionErr

	// NoChildrenErr 即没有子目录，也没有文件
	NoChildrenErr
)

func writeHash(node *Node) {
	file, err := os.Open(node.Path)
    if err != nil {
		fmt.Println(err.Error())
		displayHelpCMD()
        os.Exit(-1)
    }
    defer file.Close()
    h := sha1.New()
    _, ioErr := io.Copy(h, file)
    if ioErr != nil {
		fmt.Println(ioErr.Error())
		displayHelpCMD()
        return
	}
	
	fd, fileErr := os.OpenFile("treehash.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if fileErr != nil {

	}
	defer fd.Close()
	data  := fmt.Sprintf("%s,%x,%d\n", node.Path, h.Sum(nil), node.Size)
	
	buf := []byte(data)  
	mu.Lock()
	fd.Write(buf)
	mu.Unlock()
}

func newNode(fileInfo os.FileInfo, parentDir string) Node {
	node := Node{
		ParentDir : parentDir,
		Path      : parentDir + string(os.PathSeparator) + fileInfo.Name(),
		Name      : fileInfo.Name(),
		Size      : fileInfo.Size(),
		IsDir     : fileInfo.IsDir(),
		Hash      : "",
	}
	return node
}

// Traverse 遍历目录
func Traverse(path string, filter string) int {
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
	for i := 0; i < length; i++ {
		stack = append(stack, newNode(files[i], path))
	}
	// 树的广度优先遍历
	for len(stack) > 0 {
		node := stack[0]
		stack = stack[1:]
		if !node.IsDir {
			go writeHash(&node)
		} else {
			files, err := ioutil.ReadDir(node.Path)
			if err != nil {

			}
			if len(files) > 0 {
				for i := 0; i < len(files); i++ {
					stack = append(stack, newNode(files[i], node.Path))
				}
			}
		}
	}
	return Success
}

func displayHelpCMD() {
	fmt.Println("运行以下命令获得帮助")
	fmt.Println("go run treeash help")
}

func displayHelp() {
	fmt.Println("\n参数说明:")
	fmt.Println("-root", "要计算hash的根目录")
	fmt.Println("-filter", "需要过滤的目录或文件，支持通配符")
}

func main() {
	beginTime := time.Now()
	root      := flag.String("root", "", "root path")
	filter    := flag.String("filter", "", "filter")

	flag.Parse()

	args := flag.Args()
	if len(args) >= 1 {
		for _, value := range args {
			if value == "help" {
				displayHelp()
				break;
			}
		}
	}

	var MULTICORE int = runtime.NumCPU()
	runtime.GOMAXPROCS(MULTICORE)
	if result := Traverse(*root, *filter); result != Success {
		os.Exit(-1)
	}
	fmt.Println("duration: ", time.Now().Sub(beginTime).Seconds(), "s")
}


// 路径分隔符的问题 如 /dev/ + / + abc