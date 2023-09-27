// Code generated by hertz generator.

package main

import (
	"context"
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var wg sync.WaitGroup

func main() {
	wg.Add(2)
	ch := make(chan string)

	go func() {
		defer wg.Done()
		hertzserver(ch)

	}()

	go func() {
		defer wg.Done()
		transfile(ch)
	}()
	//todo:后台检测下载文件流程
	//go func() {
	//	defer wg.Done()
	//	getresult()
	//}()
	wg.Wait()
}

func hertzserver(ch chan string) {
	h := server.Default(server.WithHostPorts("127.0.0.1:8080"), server.WithMaxRequestBodySize(20<<20))

	h.POST("/singleFile", func(ctx context.Context, c *app.RequestContext) {
		// single file
		file, _ := c.FormFile("file")
		fmt.Println(file.Filename)

		// Upload the file to specific dst
		c.SaveUploadedFile(file, fmt.Sprintf("./file/upload/%s", file.Filename))

		c.String(consts.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})

	h.POST("/multiFile", func(ctx context.Context, c *app.RequestContext) {
		// Multipart form
		form, _ := c.MultipartForm()
		humanfiles := form.File["human"]
		tm := fmt.Sprintf("%d", time.Now().Unix())
		path := fmt.Sprintf("./file/%s", tm)
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			fmt.Println("Execute Command failed:" + err.Error())
		}
		for _, file := range humanfiles {
			fmt.Println(file.Filename)

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, path+"/human")
		}
		clothesfiles := form.File["clothes"]

		for _, file := range clothesfiles {
			fmt.Println(file.Filename)

			// Upload the file to specific dst.
			c.SaveUploadedFile(file, path+"/clothes")
		}
		//TODO:添加scp逻辑
		ch <- tm
		c.String(consts.StatusOK, fmt.Sprintf("%d files uploaded!", len(humanfiles)+len(clothesfiles)))
	})

	//h.GET("/result", func(ctx context.Context, c *app.RequestContext) {
	//	// If you use Chinese, need to encode
	//	fileName := url.QueryEscape("result")
	//	c.FileAttachment("./file/download/yumeng", fileName)
	//})
	h.GET("/result", func(ctx context.Context, c *app.RequestContext) {
		c.File("./file/download/yumeng")
	})

	h.Spin()
}

func transfile(ch chan string) {
	var path string
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		for {
			path = <-ch
			fmt.Println(path)
		}
	}()
	select {
	case <-c:
		fmt.Println("The interrupt got handled")
	}
}

func getresult() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		getfile()
	}()
	select {
	case <-c:
		fmt.Println("The interrupt got handled")
	}
}

func getfile() {
	f, err := os.Open("./file")
	if err != nil {
		fmt.Println("open file path failed:" + err.Error())
	}
	files, err := f.ReadDir(-1)
	f.Close()
	if err != nil {
		fmt.Println("search file path failed:" + err.Error())
	}
	for {
		for _, path := range files {
			_, err := os.Stat("./file/" + path.Name() + "/result")
			if err == nil {
				continue
			}
			if os.IsNotExist(err) {
				fmt.Println("find no result in path:" + path.Name())
				//尝试去远端拷贝文件
			}
		}
	}
}
