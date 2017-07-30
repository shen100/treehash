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
	ParentDir  string
	Path       string
	Name       string
	Size       int64
	isDir      bool
	Hash       string
}

func writeHash(node *Node) {
	file, err := os.Open(node.Path)
    if err != nil {
		fmt.Println(err.Error())
		displayHelp()
        return
    }
    defer file.Close()
    h := sha1.New()
    _, ioErr := io.Copy(h, file)
    if ioErr != nil {
		fmt.Println(ioErr.Error())
		displayHelp()
        return
	}
	
	fd, fileErr := os.OpenFile("treehash.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if fileErr != nil {

	}
	defer fd.Close()
	data  := fmt.Sprintf("%s,%x,%d\n", node.Path, h.Sum(nil), node.Size)
	
	buf := []byte(data)  
    fd.Write(buf)
}

func newNode(fileInfo os.FileInfo, parentDir string) Node {
	//fmt.Println("====> ", parentDir + string(os.PathSeparator) + fileInfo.Name())

	node := Node{
		ParentDir : parentDir,
		Path      : parentDir + string(os.PathSeparator) + fileInfo.Name(),
		Name      : fileInfo.Name(),
		Size      : fileInfo.Size(),
		isDir     : fileInfo.IsDir(),
		Hash      : "",
	}
	return node
}

func traverse(path string, filter string) { 
	var MULTICORE int = runtime.NumCPU()
	runtime.GOMAXPROCS(MULTICORE)
	fmt.Println("MULTICORE", MULTICORE)
	
	rootDir, err := os.Stat(path)
	
	if err != nil {
		fmt.Println(path + " 不是有效的目录")
		displayHelp()
		return
	}
	if !rootDir.IsDir() {
		fmt.Println(path + " 不是目录")
		displayHelp()
		return
	}
	var stack []Node
	files, err := ioutil.ReadDir(path)
	if err != nil {

	}
	for i := 0; i < len(files); i++ {
		stack = append(stack, newNode(files[i], path))
	}

	var readFileTime int64
	var writeFileTime int64

	for len(stack) > 0 {
		maxNum := MULTICORE
		if maxNum > len(stack) {
			maxNum = len(stack)
		}
		ch := make(chan []Node, maxNum)
		for num := 0; num < maxNum; num++ {
			go func(i int) {
				node := stack[i]
				if node.isDir {
					mu.Lock()
					beginTime := time.Now()
					files, err := ioutil.ReadDir(node.Path)
					readFileTime = readFileTime + time.Now().Sub(beginTime).Nanoseconds()
					mu.Unlock()
					if err != nil {

					}
					if len(files) > 0 {
						var nodes []Node
						for i := 0; i < len(files); i++ {
							nodes = append(nodes, newNode(files[i], node.Path))
						}
						ch <- nodes
					} else {
						ch <- nil	
					}
				} else {
					ch <- nil	
				}
			}(num)
		}
		for num := 0; num < maxNum; num++ {
			subTree := <- ch
			if (subTree != nil) {
				stack = append(stack, subTree...)	
			}
			node := stack[num]
			if !node.isDir {
				beginTime := time.Now()
				writeHash(&node)
				writeFileTime = writeFileTime + time.Now().Sub(beginTime).Nanoseconds()
			}
		}
		stack = stack[maxNum:]
		// node := stack[0]
		// stack = stack[1:]
	}
	fmt.Println("读文件总用时: ", readFileTime)
	fmt.Println("写文件总用时: ", writeFileTime)
}

// func traverse(path string, filter string) {
// 	rootDir, err := os.Stat(path)
// 	count := 0 
// 	if err != nil {
// 		fmt.Println(path + " 不是有效的目录")
// 		displayHelp()
// 		return
// 	}
// 	if !rootDir.IsDir() {
// 		fmt.Println(path + " 不是目录")
// 		displayHelp()
// 		return
// 	}
// 	var stack []Node
// 	files, err := ioutil.ReadDir(path)
// 	if err != nil {

// 	}
// 	for i := 0; i < len(files); i++ {
// 		fmt.Println(12345678, files[i].Name())
// 		stack = append(stack, newNode(files[i], path))
// 	}
// 	for len(stack) > 0 {
// 		node := stack[0]
// 		stack = stack[1:]
// 		count++
// 		if !node.isDir {
// 			writeHash(&node)
// 		} else {
// 			files, err := ioutil.ReadDir(node.Path)
// 			if err != nil {

// 			}
// 			if len(files) > 0 {
// 				for i := 0; i < len(files); i++ {
// 					stack = append(stack, newNode(files[i], node.Path))
// 				}
// 			}
// 		}
// 	}
// 	fmt.Println("总文件数: ", count)
// }

func displayHelp() {
	fmt.Println("运行以下命令获得帮助")
	fmt.Println("go run treeash help")
}

func main() {
	reqStartTime := time.Now()
	if len(os.Args) == 2 && os.Args[1] == "help" {
		fmt.Println("\n参数说明:")
		fmt.Println("-root", "目录")
		fmt.Println("-filter", "需要过滤的目录或文件，支持通配符")
		os.Exit(0)
	}

	root   := flag.String("root", "", "root path")
	filter := flag.String("filter", "", "filter")
	flag.Parse()

	fmt.Println(os.Args, *root, *filter)

	traverse(*root, *filter)

	//fmt.Println("duration: ", time.Now().Sub(reqStartTime).Seconds())
	fmt.Println("duration: ", time.Now().Sub(reqStartTime).Nanoseconds())
	
}


// 路径分隔符的问题 如 /dev/ + / + abc
// 路径是文件， 不是目录
// 路径不存在
// 目录下没有子目录或子文件