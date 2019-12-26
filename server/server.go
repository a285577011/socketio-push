package server

import (
	"context"
	"github.com/astaxie/beego"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"socketserver/conn"
	"socketserver/helper"
	"socketserver/httppush"
	"socketserver/models"
	"socketserver/models/redis"
	"strings"
	"syscall"
	"time"
)

type Server struct {
}

func NewServer() *Server {
	server := &Server{}
	return server
}
func (this *Server) SocketMiddleware(allowOrigin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		imie := c.Query("imie")
		if imie == "" {
			c.AbortWithStatus(602)
			return
		}
		//fmt.Println("IMIE"+imie)
		deviceInfo, _ := models.DeviceM.GetByImie(imie)
		if deviceInfo == nil {
			//fmt.Println("imie号不存在", deviceInfo)
			c.AbortWithStatus(601)
			return
		}
		//fmt.Println(deviceInfo)
		if deviceInfo.Ip != ip {
			c.AbortWithStatus(603)
			//return
		}
		//c.Writer.Header().Set("Access-Control-Allow-Origin", "http://192.168.66.166:9090")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, Content-Length, X-CSRF-Token, Token, session, Origin, Host, Connection, Accept-Encoding, Accept-Language, X-Requested-With")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Request.Header.Del("Origin")

		c.Next()
	}
}
func (this *Server) HttpMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		beego.LoadAppConfig("ini", "../conf/common.conf")
		key := beego.AppConfig.String("allowPushKey")
		header := c.Request.Header
		token, ok := header["Token"]
		if !ok {
			c.AbortWithStatus(601)
			return
		}
		if key != token[0] {
			c.AbortWithStatus(602)
			return
		}
		allowIp := beego.AppConfig.String("allowPushIp")
		if allowIp == "" {
			c.AbortWithStatus(603)
			return
		}
		allowIpSlice := strings.Split(allowIp, ",")
		if !helper.Contains(allowIpSlice, ip) {
			c.AbortWithStatus(403)
			return
		}
		c.Request.Header.Del("Origin")

		c.Next()
	}
}
func (this *Server) Run() {
	router := gin.New()
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.OnConnect("/", func(s socketio.Conn) error {
		//s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})
	server.OnEvent("/", "join", func(s socketio.Conn, imie string, group string) {
		if group != "" {
			s.Join(group)
		}
		context := map[string]string{s.ID(): imie}
		s.SetContext(context)
		conn.ImieMap.Store(imie, s.ID())
		conn.ConnectionMap.Store(s.ID(), s)
		s.Join(conn.GetDefaultGroup()) //全局分组
		//s.Join(s.ID())
		//s.Emit(conn.GetDefaultEvent(), s.ID())//默认事件推送
		log.Println("join:imie-" + imie)
		//fmt.Println(conn.ConnectionMap)
	})
	server.OnEvent("/", "task-reply", func(s socketio.Conn, msgId string) {
		if msgId == "" {
			log.Println("msgID参数缺失")
			return
		}
		context := s.Context()
		switch context.(type) {
		case map[string]string:
			contextMap := context.(map[string]string)
			imie := contextMap[s.ID()]
			//fmt.Println("task-reply",imie)
			redis.RedisModel.Hdel("SinglePush-imie:"+imie, msgId) //消息确认接收 删除记录
		}
	})
	/*server.OnEvent("/", "bye", func(s socketio.Conn) string {
		fmt.Println("aaaa")
		last := s.Context().(string)
		s.Emit("bye", last)
		s.LeaveAll()
		//conn.ImieMap.Delete(s.ID())
		conn.ConnectionMap.Delete(s.ID())
		s.Close()
		return last
	})*/
	server.OnError("/", func(s socketio.Conn, e error) {
		now := time.Now()
		log.Println(now.Format("2006-01-02 15:04:05")+" meet error:", e)
	})
	server.OnDisconnect("/", func(s socketio.Conn, msg string) {
		context := s.Context()
		switch context.(type) {
		case map[string]string:
			contextMap := context.(map[string]string)
			conn.ImieMap.Delete(contextMap[s.ID()])
			log.Println("disconnect:imie-", contextMap[s.ID()])
		}
		conn.ConnectionMap.Delete(s.ID())
		log.Println("disconnect:connId-", s.ID())
		s.LeaveAll()
		s.Close()
	})

	go server.Serve()
	defer server.Close()
	httpPush := httppush.NewPush(server)
	router.POST("/grouppush/*any", this.HttpMiddleware(), httpPush.GroupPush)
	router.POST("/singlepush/*any", this.HttpMiddleware(), httpPush.SinglePush)
	router.POST("/imieisconn/*any", this.HttpMiddleware(), httpPush.CheckIsConn)
	router.Use(this.SocketMiddleware(""))
	{
		router.GET("/socket.io/*any", gin.WrapH(server))
		router.POST("/socket.io/*any", gin.WrapH(server))
	}
	log.Println("start...")
	//router.StaticFS("/public", http.Dir("../asset"))
	srv := &http.Server{
		Addr:    ":5050",
		Handler: router,
	}
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	//router.Run(":5050")
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
