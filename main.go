package main

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

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
		fmt.Println("???", err.Error())
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
    //fmt.Printf("%s,%x,%d\n", node.Path, h.Sum(nil), node.Size)
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
	rootDir, err := os.Stat(path)
	count := 0 
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
		fmt.Println(12345678, files[i].Name())
		stack = append(stack, newNode(files[i], path))
	}
	for len(stack) > 0 {
		node := stack[0]
		stack = stack[1:]
		count++
		if !node.isDir {
			writeHash(&node)
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
	fmt.Println("总文件数: ", count)
}

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

	fmt.Println("duration: ", time.Now().Sub(reqStartTime).Seconds())
	fmt.Println("duration: ", time.Now().Sub(reqStartTime).Nanoseconds())
	
}


// 路径分隔符的问题 如 /dev/ + / + abc
// 路径是文件， 不是目录
// 路径不存在