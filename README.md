# treehash
一个生成目录树哈希的小工具  

## 示例
进入项目目录下，然后运行以下命令

```
go build  
./treehash -root=/Users/liushen/dev/docs
```  

![image](https://github.com/shen100/treehash/raw/master/images/1.png)    


## 命令与参数
运行`./treehash help`，显示帮助  
![image](https://github.com/shen100/treehash/raw/master/images/2.png)  

| 参数 | 说明 |
| -------- | -------- |
| -root     | 要生成hash的根目录     |
| -filter   | 过滤目录或文件，支持通配符     |
| -output   | 最后写入的文件路径     |