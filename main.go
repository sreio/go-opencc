package main

import (
	"embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/longbridgeapp/opencc"
	flag "github.com/spf13/pflag"
	"log"
	"net/http"
)

type RequestData struct {
	Text string `form:"text" json:"text"`
}

var S2T *opencc.OpenCC

func init() {
	var err error
	S2T, err = opencc.New("s2twp")
	if err != nil {
		log.Fatal(err)
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

var Port int

//go:embed index.html
var indexHtml embed.FS

//go:embed api.html
var ApiHtml embed.FS

func main() {
	flag.IntVarP(&Port, "port", "p", 8581, "port")
	flag.Parse()

	r := gin.Default()
	r.Use(Cors())
	r.GET("/", func(c *gin.Context) {
		data, err := indexHtml.ReadFile("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "加载页面出错")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// 文档
	r.GET("/docs", func(c *gin.Context) {
		data, err := ApiHtml.ReadFile("api.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "加载文档失败：%v", err)
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	r.GET("/ping", ping)
	r.GET("/s2twp", getS2twp)
	r.POST("/s2twp", postS2twp)
	r.GET("/types", typeList)
	r.POST("/convert", convert)

	err := http.ListenAndServe(fmt.Sprintf(":%d", Port), r)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func getS2twp(c *gin.Context) {
	text := c.DefaultQuery("text", "")
	if text == "" {
		c.JSON(200, gin.H{
			"text": "",
		})
		return
	}

	out, err := S2T.Convert(text)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(200, gin.H{
		"text": out,
	})
}

func postS2twp(c *gin.Context) {
	var req RequestData
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(200, gin.H{
			"text": "",
		})
		return
	}

	out, err := S2T.Convert(req.Text)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(200, gin.H{
		"text": out,
	})
}

func convert(c *gin.Context) {
	var req struct {
		Text string `form:"text" json:"text"`
		Type string `form:"type" json:"type"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(200, gin.H{
			"text": "",
			"msg":  err.Error(),
		})
		return
	}

	if req.Type == "" || req.Text == "" {
		c.JSON(200, gin.H{
			"text": "",
			"msg":  "缺少参数",
		})
		return
	}

	opencc, err := opencc.New(req.Type)
	if err != nil {
		log.Fatal(err)
	}

	out, err := opencc.Convert(req.Text)
	if err != nil {
		log.Fatal(err)
	}

	c.JSON(200, gin.H{
		"text": out,
	})
}

func typeList(c *gin.Context) {
	// s2t, t2s, s2tw, tw2s, s2hk, hk2s, s2twp, tw2sp, t2tw, t2hk, t2jp, jp2t, tw2t, s2hk-finance
	c.JSON(200, gin.H{
		"list": map[string]string{
			"s2t":          "简转繁",
			"t2s":          "繁转简",
			"s2tw":         "简转繁(台湾繁体)",
			"tw2s":         "繁(台湾繁体)转简",
			"s2hk":         "简转繁(香港繁体)",
			"hk2s":         "繁(香港繁体)转简",
			"s2twp":        "简转繁(台湾标准-台湾常用词汇)",
			"tw2sp":        "繁(香港繁体)转简(中国大陆常用词汇)",
			"t2tw":         "繁转繁(台湾繁体)",
			"t2hk":         "繁转繁(香港繁体)",
			"t2jp":         "繁转繁(日文新字体)",
			"jp2t":         "繁(日文新字体)转繁",
			"tw2t":         "繁(台湾繁体)转繁",
			"s2hk-finance": "简转繁(香港繁体-金融)",
		},
	})
}
