// Copyright 2022 Ze-Bin Wang.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package tokenizer

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var dictionary = &Dictionary{
	dict: make(map[string]int),
}

type Dictionary struct {
	dict map[string]int
	mu   sync.RWMutex
	tf   int // total freq
}

func (d *Dictionary) Exist(word string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.dict[strings.ToLower(word)]
	return ok
}

func (d *Dictionary) GetWord(word string) (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	v, ok := d.dict[strings.ToLower(word)]
	return v, ok
}

func (d *Dictionary) GetTotalFreq() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return float64(d.tf)
}

func (d *Dictionary) AddWord(
	word string,
	freq int,
	prop string,
) (exist bool, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if _, ok := d.dict[strings.ToLower(word)]; ok {
		exist = true
		return
	}

	dictStdFile, err := GetDictFile(DictStdFile)
	if err != nil {
		return
	}

	dictUserFile := filepath.Dir(dictStdFile)
	dictUserFile += string(filepath.Separator) + DictUserFile

	f, err := os.OpenFile(
		dictUserFile,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0666,
	)
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()

	stat, err := f.Stat()
	if err != nil {
		return
	}

	line := ""
	n := stat.Size()
	if n > 0 {
		buf := make([]byte, 1)
		_, err = f.ReadAt(buf, n-1)
		if err != nil {
			return
		}
		if buf[0] != '\n' {
			line += "\n"
		}
	}
	line += word + " " + strconv.Itoa(freq) + " " + prop + "\n"
	_, err = f.Write([]byte(line))
	if err != nil {
		log.Println(err)
		return
	}

	d.dict[strings.ToLower(word)] = freq
	d.tf += freq
	return
}

func (d *Dictionary) load(fileDict string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	f, err := os.Open(fileDict)
	if err != nil {
		log.Println(err)
		return errors.New(
			"unable to load the dictionary library:" + filepath.Base(fileDict),
		)
	}
	defer f.Close()
	return d.loadFromReader(f)
}

func (d *Dictionary) loadFromReader(f io.Reader) error {
	itemCount := 0
	reader := bufio.NewReader(f)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}

		elem := strings.Fields(line)
		if len(elem) != 3 {
			if err == io.EOF {
				break
			}
			continue
		}

		itemCount++
		nFreq, err := strconv.Atoi(elem[1])
		if err != nil {
			nFreq = 0
		}
		d.tf += nFreq
		d.dict[strings.ToLower(elem[0])] = nFreq

		runeWord := []rune(elem[0])
		for i := range runeWord {
			s := strings.ToLower(string(runeWord[:i+1]))
			if _, ok := d.dict[s]; !ok {
				d.dict[s] = 0
			}
		}

		if err == io.EOF {
			break
		}
	}
	if len(d.dict) == 0 || d.tf <= 0 {
		return errors.New(
			"unable to load the dictionary, dictionary is empty",
		)
	}
	return nil
}

func InitDictionary() {
	// load the standard dictionary
	if dictFS == nil {
		dictStdFile, err := GetDictFile(DictStdFile)
		if err != nil {
			log.Panicln(err)
		}
		err = dictionary.load(dictStdFile)
		if err != nil {
			log.Panicln(err)
		}

		// load the user-defined dictionary
		dictUserFile, err := GetDictFile(DictUserFile)
		if err == nil {
			dictionary.load(dictUserFile)
		}
	} else {
		content, err := fs.ReadFile(dictFS, DictStdFile)
		if err != nil {
			log.Panicln(err)
		}
		f := bytes.NewReader(content)
		err = dictionary.loadFromReader(f)
		if err != nil {
			log.Panicln(err)
		}

		content, err = fs.ReadFile(dictFS, DictUserFile)
		if err == nil {
			// only load user dict when it exists
			f = bytes.NewReader(content)
			err = dictionary.loadFromReader(f)
			if err != nil {
				log.Panicln(err)
			}
		}
	}
}

func GetDictionary() *Dictionary {
	return dictionary
}
