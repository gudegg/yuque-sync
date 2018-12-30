package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

var REGEX = "decodeURIComponent\\(\"(.+)\"\\)\\);"

var INDEX_URL = "https://www.yuque.com/%v"
var DOC_URL = "https://www.yuque.com/api/docs/%v?book_id=%v&mode=markdown"

const (
	DOC   string = "DOC"
	TITLE string = "TITLE"
	HUGO  string = "hugo" //hugo格式
	HEXO  string = "hexo" //hexo格式
	RAW   string = "raw"  //原始格式
)

var TimeFormat = "2006-01-02 15:04:05"

func main() {
	n := flag.String("n", "", "设置namespace")
	p := flag.String("p", "download/", "设置存储路径")
	t := flag.String("t", RAW, "文件写入格式,支持hexo、hugo")
	o := flag.Bool("o", true, "文件同名覆盖写入")
	flag.Parse()
	//namespace
	*n = strings.TrimSpace(*n)
	*n = strings.TrimPrefix(*n, "/")
	*n = strings.TrimSuffix(*n, "/")
	var namespace = *n
	if len(namespace) == 0 {
		panic("namespace不能为空,请使用-n设置")
	}

	//filePath存储路径
	*p = strings.TrimSpace(*p)
	var filePath = *p
	if len(filePath) == 0 {
		panic("存储路径不能为空,请使用-p设置")
	}
	if !strings.HasSuffix(filePath, "/") {
		filePath = filePath + "/"
	}
	//type写入文件格式
	*t = strings.TrimSpace(*t)
	var writeType = *t
	writeType = strings.ToLower(writeType)
	if len(writeType) == 0 {
		panic("写入文件格式不能为空,请使用-t设置")
	}

	//overwrite覆盖文件写入
	var overwrite = *o

	index := fmt.Sprintf(INDEX_URL, namespace)
	log("当前设置主页地址:", index, ",namespace:", namespace, ",写入路径:", filePath, ",类型:", writeType, ",文件已存在覆盖:", overwrite)
	html := httpGet(index)
	json := getNamespaceData(html)
	jsonObject := gjson.Parse(json)
	bookId := jsonObject.Get("book.id").String()
	var tag string
	log("开始语雀文档下载")
	jsonObject.Get("book.toc").ForEach(func(key, value gjson.Result) bool {
		tp := value.Get("type").String()
		switch tp {
		case DOC:
			urlResult := value.Get("url")
			url := fmt.Sprintf(DOC_URL, urlResult.String(), bookId)
			downloadAndWrite(url, tag, writeType, filePath, overwrite, value)
			break
		case TITLE:
			tag = value.Get("title").Str
			log("tag:", tag)
			break
		}
		return true
	})
	log("完成语雀文档下载")

}

var client = &http.Client{}

func httpGet(url string) string {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "my-app-name")
	//限制下请求速度
	time.Sleep(time.Millisecond * 200)
	resp, err := client.Do(request)
	check(err)
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	check(err)
	return string(bytes)
}

//获取空间首页数据
func getNamespaceData(html string) string {
	r, _ := regexp.Compile(REGEX)
	match := r.FindStringSubmatch(html)
	if len(match) < 1 {
		panic("未获取到空间首页数据")
	} else {
		decodeStr, err := url.QueryUnescape(match[1])
		check(err)
		return decodeStr
	}
}

func downloadAndWrite(url string, tag string, writeType string, filePath string, overwrite bool, value gjson.Result) {
	result := httpGet(url)
	sourceCode := gjson.Get(result, "data.sourcecode").Str
	title := value.Get("title").String()
	createTime := gjson.Get(result, "data.created_at").Time()
	fileName := title + ".md"
	//替换文件名包含斜杠
	fileName = strings.Replace(fileName, "/", "", -1)
	var buffer bytes.Buffer
	if strings.EqualFold(HEXO, writeType) || strings.EqualFold(HUGO, writeType) {
		buffer.WriteString("---\n")
		buffer.WriteString("title: " + title + "\n")
		buffer.WriteString("date: " + createTime.Local().Format(TimeFormat) + "\n")
		buffer.WriteString("tags: [" + tag + "]\n")
		buffer.WriteString("---\n")
		//去除第一行标题
		split := strings.Split(sourceCode, "\n")
		var tempSlice = split[1:]
		sourceCode = strings.Join(tempSlice, "\n")
	}
	buffer.WriteString(sourceCode)
	if !fileExists(filePath) {
		os.MkdirAll(filePath, 0700)
	}
	//不覆盖 已经存在文件跳过
	if !overwrite {
		if fileExists(filePath + fileName) {
			log("当前文档已存在忽略写入:", fileName)
			return
		}
	}
	file, e := os.Create(filePath + fileName)
	check(e)
	defer file.Close()
	log("当前正在写入文档:", fileName)
	writer := bufio.NewWriter(file)
	writer.Write(buffer.Bytes())
	writer.Flush()
	file.Sync()
}

func log(content ...interface{}) {
	fmt.Println(content)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
