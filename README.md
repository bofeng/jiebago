JiebaGo 是 jieba 中文分词的 Go 语言版本。

本仓库是在源仓库 https://github.com/wangshizebin/jiebago 的基础上，添加了可以通过`fs.FS`来初始化一个jiebago的实例。详细请参考下面的"四种方法可以初始化"的部分。

## 功能特点

+ 支持多种分词方式，包括: 最大概率模式, HMM新词发现模式, 搜索引擎模式, 全模式
+ 支持抽取关键词，包括: 无权重关键词, 权重关键词
+ 支持多种使用方式，包括: Go语言包, Windows Dll, Web API, Docker
+ 支持在线并行添加字典词库和停止词
+ 全部代码使用 go 语言实现，全面兼容 jieba python 词库

## 引用方法

```bash
go get github.com/bofeng/jiebago
```

## 使用

由于分词和提取关键词使用了中文预置词库和TF-IDF统计库，所以使用 jiebago，需要先下载项目中词库`dictionary`目录，并将`dictionary`放入项目的工作目录中。

在下载本仓库的`dictionary`目录到你自己的仓库后，有四种方法可以初始化`jiebago` instance:

1，使用默认文件路径，此时jiebago会自动在当前目录寻找`dictionary`目录：

```go
jieBaGo := jiebago.NewJieBaGo()
```

2，提供dictionary目录的路径：

```go
jieBaGo := jiebago.NewJieBaGo("path/to/dictionary/folder")
```

3，将`dictionary`目录封装成`fs.FS`，然后调用`NewJieBaGoWithFS`的方法初始化：

```go
jieBaGo := jiebago.NewJieBaGoWithFS(os.DirFS("dictionary"))
```

4，同3，但使用embed.FS将dictionary目录整个打包进编译后的文件。这样的好处是，当发布你的binary时，不需要再另外包含`dictionary`目录及其文件，因为已经将它们一同打包进了编译后的binary文件。

```go
//go:embed dictionary
var embedFS embed.FS

dictFS, err := fs.Sub(embedFS, "dictionary")
if err != nil {
	log.Panicln(err)
}
jieBaGo := jiebago.NewJieBaGoWithFS(dictFS)
```

## 功能示例

```golang
package main

import (
	"fmt"
	"strings"

	"github.com/bofeng/jiebago"
)

func main() {
	jieBaGo := jiebago.NewJieBaGo()
	// 可以指定字典库的位置
	// jieBaGo := jiebago.NewJieBaGo("/data/mydict")

	sentence := "Shell 位于用户与系统之间，用来帮助用户与操作系统进行沟通。通常都是文字模式的 Shell。"
	fmt.Println("原始语句：", sentence)
	fmt.Println()

	// 默认模式分词
	words := jieBaGo.Cut(sentence)
	fmt.Println("默认模式分词：", strings.Join(words,"/"))

	// 精确模式分词
	words = jieBaGo.CutAccurate(sentence)
	fmt.Println("精确模式分词：", strings.Join(words,"/"))

	// 全模式分词
	words = jieBaGo.CutFull(sentence)
	fmt.Println("全模式分词：", strings.Join(words,"/"))

	// NoHMM模式分词
	words = jieBaGo.CutNoHMM(sentence)
	fmt.Println("NoHMM模式分词：", strings.Join(words,"/"))

	// 搜索引擎模式分词
	words = jieBaGo.CutForSearch(sentence)
	fmt.Println("搜索引擎模式分词：", strings.Join(words,"/"))
	fmt.Println()

	// 提取关键词，即Tag标签
	keywords := jieBaGo.ExtractKeywords(sentence, 20)
	fmt.Println("提取关键词：", strings.Join(keywords,"/"))

	// 提取带权重的关键词，即Tag标签
	keywordsWeight := jieBaGo.ExtractKeywordsWeight(sentence, 20)
	fmt.Println("提取带权重的关键词：", keywordsWeight)
	fmt.Println()

	// 向字典加入单词
	exist, err := jieBaGo.AddDictWord("编程宝库", 3, "n")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("向字典加入单词：编程宝库")
		if exist {
			fmt.Println("单词已经存在")
		}
	}

	// 向字典加入停止词
	exist, err = jieBaGo.AddStopWord("the")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("向字典加入停止词：the")
		if exist {
			fmt.Println("单词已经存在")
		}
	}
}
```

```
原始语句： Shell位于用户与系统之间，用来帮助用户与操作系统进行沟通。

默认模式分词： Shell/位于/用户/与/系统/之间/，/用来/帮助/用户/与/操作系统/进行/沟通/。
精确模式分词： Shell/位于/用户/与/系统/之间/，/用来/帮助/用户/与/操作系统/进行/沟通/。
全模式分词： Shell/位于/用户/与/系统/之间/，/用来/帮助/用户/与/操作/操作系统/系统/进行/沟通/。
NoHMM模式分词： Shell/位于/用户/与/系统/之间/，/用来/帮助/用户/与/操作系统/进行/沟通/。
搜索引擎模式分词： Shell/位于/用户/与/系统/之间/，/用来/帮助/用户/与/操作/系统/操作系/操作系统/进行/沟通/。

提取关键词： 用户/Shell/操作系统/沟通/帮助/位于/系统/之间/进行
提取带权重的关键词： [{用户 1.364467214484} {Shell 1.19547675029} {操作系统 0.9265948663750001} {沟通 0.694890548758} {帮助 0.5809050240370001} {位于 0.496609078159} {系统 0.49601794343199995} {之间 0.446152979906} {进行 0.372712479502}]

向字典加入单词：编程宝库
向字典加入停止词：the
```

更详细的例子参照 example/main.go, api/iebago_test.go 中的代码。

## 单元测试
go 包

```bash
go test
```

Web API

```bash
cd api
go test 
```