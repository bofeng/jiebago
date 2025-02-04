// Copyright 2022 Ze-Bin Wang.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/bofeng/jiebago"
	"github.com/bofeng/jiebago/tokenizer"
	"github.com/gin-gonic/gin"
)

const (
	Success = iota
	ErrorFail
	ErrorRequestMethod
	ErrorJsonData
	ErrorWordEmpty
	ErrorWeightInteger
	ErrorWeightRange
	ErrorCountInteger
)

var (
	jieBaGo = &jiebago.JieBaGo{}
)

func main() {
	httpAddr := flag.String(
		"http_addr",
		":8118",
		"http_addr specifies the listening ip and port, for example: -http_addr 1.2.3.4:8888",
	)

	dictPath := flag.String(
		"dict_path",
		"",
		"dict_path specifies the path of dictionary, for example: -dict_path /data/dictionary",
	)

	flag.Parse()

	jieBaGo = jiebago.NewJieBaGo(*dictPath)

	engine := gin.Default()

	engine.Any("/cut_words", cutWordsHandler)
	engine.Any("/extract_keywords", extractKeywordsHandler)
	engine.Any("/add_dict_word", addDictWordHandler)
	engine.Any("/add_stop_word", addStopWordHandler)

	if err := engine.Run(*httpAddr); err != nil {
		log.Print(err)
	}
}

type RequestCutWord struct {
	Sentence string `json:"s"`
	Mode     string `json:"mode"`
}

type RequestExtractWord struct {
	Sentence string `json:"s"`
	Mode     string `json:"mode"`
	Count    int    `json:"count"`
}

type RequestAddWord struct {
	Word   string `json:"s"`
	Weight int    `json:"weight"`
	Prop   string `json:"prop"`
}

type Response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func cutWordsHandler(c *gin.Context) {
	sentence := ""
	mode := ""
	if c.Request.Method == "GET" {
		mode = strings.ToLower(c.DefaultQuery("mode", ""))
		sentence = c.DefaultQuery("s", "")
	} else if c.Request.Method == "POST" {
		var request RequestCutWord
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusOK, struct {
				Response
				Words []string `json:"words"`
			}{
				Response: Response{
					ErrCode: ErrorJsonData,
					ErrMsg:  `invalid json data, the proper data format is {"s":"xx","mode":"xx"}`,
				},
				Words: []string{},
			})
			return
		}
		mode = request.Mode
		sentence = request.Sentence
	} else {
		c.JSON(http.StatusOK, struct {
			Response
			Words []string `json:"words"`
		}{
			Response: Response{
				ErrCode: ErrorRequestMethod,
				ErrMsg:  `invalid request method, only GET and POST methods are supported`,
			},
			Words: []string{},
		})
		return
	}

	var words []string
	if mode == "full" {
		words = jieBaGo.CutFull(sentence)
	} else if mode == "accurate" {
		words = jieBaGo.CutAccurate(sentence)
	} else if mode == "nohmm" {
		words = jieBaGo.CutNoHMM(sentence)
	} else if mode == "search" {
		words = jieBaGo.CutForSearch(sentence)
	} else {
		words = jieBaGo.Cut(sentence)
	}

	c.JSON(http.StatusOK, struct {
		Response
		Words []string `json:"words"`
	}{
		Response: Response{
			ErrCode: Success,
			ErrMsg:  "success",
		},
		Words: words,
	})
}

func extractKeywordsHandler(c *gin.Context) {
	sentence := ""
	count := 0
	mode := ""
	if c.Request.Method == "GET" {
		sentence = c.DefaultQuery("s", "")
		mode = c.DefaultQuery("mode", "")
		w := c.DefaultQuery("count", "0")
		var err error
		count, err = strconv.Atoi(w)
		if err != nil {
			c.JSON(http.StatusOK, struct {
				Response
				Tags []string `json:"tags"`
			}{
				Response: Response{
					ErrCode: ErrorCountInteger,
					ErrMsg:  "the count must be an integer",
				},
				Tags: []string{},
			})
			return
		}
	} else if c.Request.Method == "POST" {
		var request RequestExtractWord
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusOK, struct {
				Response
				Tags []string `json:"tags"`
			}{
				Response: Response{
					ErrCode: ErrorJsonData,
					ErrMsg:  `invalid json data, the proper data format is {"s":"xx","count":xx,"mode":"xx"}`,
				},
				Tags: []string{},
			})
			return
		}
		sentence = request.Sentence
		mode = request.Mode
		count = request.Count
	} else {
		c.JSON(http.StatusOK, struct {
			Response
			Tags []string `json:"tags"`
		}{
			Response: Response{
				ErrCode: ErrorRequestMethod,
				ErrMsg:  `invalid request method, only GET and POST methods are supported`,
			},
			Tags: []string{},
		})
		return
	}
	if count <= 0 {
		count = 20
	}

	if mode == "weight" {
		tags := jieBaGo.ExtractKeywordsWeight(sentence, count)
		c.JSON(http.StatusOK, struct {
			Response
			Tags []tokenizer.Keyword `json:"tags"`
		}{
			Response: Response{
				ErrCode: Success,
				ErrMsg:  "success",
			},
			Tags: tags,
		})
	} else {
		tags := jieBaGo.ExtractKeywords(sentence, count)
		c.JSON(http.StatusOK, struct {
			Response
			Tags []string `json:"tags"`
		}{
			Response: Response{
				ErrCode: Success,
				ErrMsg:  "success",
			},
			Tags: tags,
		})
	}
}

func addDictWordHandler(c *gin.Context) {
	word := ""
	weight := 0
	prop := ""
	if c.Request.Method == "GET" {
		word = c.DefaultQuery("s", "")
		w := c.DefaultQuery("weight", "0")
		var err error
		weight, err = strconv.Atoi(w)
		if err != nil {
			c.JSON(http.StatusOK, Response{
				ErrCode: ErrorWeightInteger,
				ErrMsg:  "the weight must be an integer",
			})
			return
		}
		prop = c.DefaultQuery("prop", "")
	} else if c.Request.Method == "POST" {
		var request RequestAddWord
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusOK, Response{
				ErrCode: ErrorJsonData,
				ErrMsg:  `invalid json data, the proper data format is {"s":"xx","weight":xx,"prop":"xx"}`,
			})
			return
		}
		word = request.Word
		weight = request.Weight
		prop = request.Prop
	}

	word = strings.TrimSpace(word)
	if len(word) == 0 {
		c.JSON(http.StatusOK, Response{
			ErrCode: ErrorWordEmpty,
			ErrMsg:  "the word is empty",
		})
		return
	}

	if weight < 0 || weight > 5000 {
		c.JSON(http.StatusOK, Response{
			ErrCode: ErrorWeightRange,
			ErrMsg:  "the weight must be between 0 and 5000",
		})
		return
	}

	if prop == "" {
		prop = "n"
	}

	exist, err := jieBaGo.AddDictWord(word, weight, prop)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			ErrCode: ErrorFail,
			ErrMsg:  err.Error(),
		})
		return
	}

	message := "success"
	if exist {
		message = "the word already exists"
	}
	c.JSON(http.StatusOK, Response{
		ErrCode: Success,
		ErrMsg:  message,
	})
}

func addStopWordHandler(c *gin.Context) {
	word := ""
	if c.Request.Method == "GET" {
		word = c.DefaultQuery("s", "")
	} else if c.Request.Method == "POST" {
		var request RequestAddWord
		err := c.BindJSON(&request)
		if err != nil {
			c.JSON(http.StatusOK, Response{
				ErrCode: ErrorJsonData,
				ErrMsg:  `invalid json data, the proper data format is {"s":"xx"}`,
			})
			return
		}
		word = request.Word
	}

	word = strings.TrimSpace(word)
	if len(word) == 0 {
		c.JSON(http.StatusOK, Response{
			ErrCode: ErrorWordEmpty,
			ErrMsg:  "the word is empty",
		})
		return
	}

	exist, err := jieBaGo.AddStopWord(word)
	if err != nil {
		c.JSON(http.StatusOK, Response{
			ErrCode: ErrorFail,
			ErrMsg:  err.Error(),
		})
		return
	}

	message := "success"
	if exist {
		message = "the word already exists"
	}
	c.JSON(http.StatusOK, Response{
		ErrCode: Success,
		ErrMsg:  message,
	})
}
