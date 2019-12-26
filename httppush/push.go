package httppush

import (
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	conn2 "socketserver/conn"
	"socketserver/models/redis"
	"strconv"
	"time"

	//"net/http"
)

type Push struct {
	ioServer *socketio.Server

}
func NewPush(Server *socketio.Server) *Push{
	p:=&Push{ioServer:Server}
	return p
}
func (this *Push) GroupPush(c *gin.Context) {
	msg := c.PostForm("msg")
	group := c.PostForm("group")
	if group==""{
		group=conn2.GetDefaultGroup()
	}
	len:=this.ioServer.RoomLen(group)
	if len==0{
		c.JSON(200,gin.H{
			"code":1001,
			"msg":"分组无连接",
		})
		return
	}
	this.ioServer.BroadcastToRoom(group,conn2.GetDefaultEvent(),msg)
	c.JSON(200,gin.H{
		"code":0,
		"msg":"发送成功",
	})
	return
}
func (this *Push) SinglePush(c *gin.Context) {
	imie := c.PostForm("imie")
	pushId,ok1:=conn2.ImieMap.Load(imie)
	if !ok1{
		c.JSON(200,gin.H{
			"code":202,
			"msg":"imie未连接",
		})
		return
	}
	msg := c.PostForm("msg")
	pushIdS:=pushId.(string)
	conn,ok :=conn2.ConnectionMap.Load(pushIdS)
	if !ok{
		c.JSON(200,gin.H{
			"code":201,
			"msg":"连接不存在",
		})
		return
	}else{
		msgId:=this.GetMsgId()
		cs := conn.(socketio.Conn)
		cs.Emit(conn2.GetDefaultEvent(),msg,msgId)
		now := time.Now()
		redis.RedisModel.Hset("SinglePush-imie:"+imie,msgId,"time:"+now.Format("2006-01-02 15:04:05")+";msg:"+msg)
	}
	c.JSON(200,gin.H{
		"code":0,
		"msg":"发送成功",
	})
	return
}
func (this *Push) GetMsgId() string{
	key:="pushMsgId"
	id,_:=redis.RedisModel.Incr(key)
	idstr:=strconv.Itoa(id)
	return idstr
}
func (this *Push) CheckIsConn(c *gin.Context) {
	imie := c.PostForm("imie")
	_,ok1:=conn2.ImieMap.Load(imie)
	if !ok1{
		c.JSON(200,gin.H{
			"code":202,
			"msg":"imie未连接",
		})
		return
	}
	c.JSON(200,gin.H{
		"code":0,
		"msg":"已连接",
	})
	return
}