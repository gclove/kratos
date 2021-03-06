package main

const (
	_tplAppToml = `
# This is a TOML document. Boom~
`

	_tplMySQLToml = `
[demo]
	addr = "127.0.0.1:3306"
	dsn = "{user}:{password}@tcp(127.0.0.1:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8"
	readDSN = ["{user}:{password}@tcp(127.0.0.2:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8mb4,utf8","{user}:{password}@tcp(127.0.0.3:3306)/{database}?timeout=1s&readTimeout=1s&writeTimeout=1s&parseTime=true&loc=Local&charset=utf8,utf8mb4"]
	active = 20
	idle = 10
	idleTimeout ="4h"
	queryTimeout = "200ms"
	execTimeout = "300ms"
	tranTimeout = "400ms"
`
	_tplMCToml = `
demoExpire = "24h"

[demo]
	name = "{{.Name}}"
	proto = "tcp"
	addr = "127.0.0.1:11211"
	active = 50
	idle = 10
	dialTimeout = "100ms"
	readTimeout = "200ms"
	writeTimeout = "300ms"
	idleTimeout = "80s"
`
	_tplRedisToml = `
demoExpire = "24h"

[demo]
	name = "{{.Name}}"
	proto = "tcp"
	addr = "127.0.0.1:6389"
	idle = 10
	active = 10
	dialTimeout = "1s"
	readTimeout = "1s"
	writeTimeout = "1s"
	idleTimeout = "10s"
`

	_tplHTTPToml = `
[server]
    addr = "0.0.0.0:8000"
    timeout = "1s"
`
	_tplGRPCToml = `
[server]
    addr = "0.0.0.0:9000"
    timeout = "1s"
`

	_tplChangeLog = `## {{.Name}}

### v1.0.0
1. 上线功能xxx
`
	_tplMain = `package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"{{.Name}}/internal/server/grpc"
	"{{.Name}}/internal/server/http"
	"{{.Name}}/internal/service"
	"github.com/bilibili/kratos/pkg/conf/paladin"
	"github.com/bilibili/kratos/pkg/log"
)

func main() {
	flag.Parse()
	if err := paladin.Init(); err != nil {
		panic(err)
	}
	log.Init(nil) // debug flag: log.dir={path}
	defer log.Close()
	log.Info("{{.Name}} start")
	svc := service.New()
	grpcSrv := grpc.New(svc)
	httpSrv := http.New(svc)
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Info("get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
			defer cancel()
			grpcSrv.Shutdown(ctx)
			httpSrv.Shutdown(ctx)
			log.Info("{{.Name}} exit")
			svc.Close()
			time.Sleep(time.Second)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
`

	_tplContributors = `# Owner
{{.Owner}}

# Author

# Reviewer
`
	_tplDao = `package dao

import (
	"context"
	"time"

	"github.com/bilibili/kratos/pkg/cache/memcache"
	"github.com/bilibili/kratos/pkg/cache/redis"
	"github.com/bilibili/kratos/pkg/conf/paladin"
	"github.com/bilibili/kratos/pkg/database/sql"
	"github.com/bilibili/kratos/pkg/log"
	xtime "github.com/bilibili/kratos/pkg/time"
)

// Dao dao.
type Dao struct {
	db          *sql.DB
	redis       *redis.Pool
	redisExpire int32
	mc          *memcache.Memcache
	mcExpire    int32
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// New new a dao and return.
func New() (dao *Dao) {
	var (
		dc struct {
			Demo *sql.Config
		}
		rc struct {
			Demo       *redis.Config
			DemoExpire xtime.Duration
		}
		mc struct {
			Demo       *memcache.Config
			DemoExpire xtime.Duration
		}
	)
	checkErr(paladin.Get("mysql.toml").UnmarshalTOML(&dc))
	checkErr(paladin.Get("redis.toml").UnmarshalTOML(&rc))
	checkErr(paladin.Get("memcache.toml").UnmarshalTOML(&mc))
	dao = &Dao{
		// mysql
		db: sql.NewMySQL(dc.Demo),
		// redis
		redis:       redis.NewPool(rc.Demo),
		redisExpire: int32(time.Duration(rc.DemoExpire) / time.Second),
		// memcache
		mc:       memcache.New(mc.Demo),
		mcExpire: int32(time.Duration(mc.DemoExpire) / time.Second),
	}
	return
}

// Close close the resource.
func (d *Dao) Close() {
	d.mc.Close()
	d.redis.Close()
	d.db.Close()
}

// Ping ping the resource.
func (d *Dao) Ping(ctx context.Context) (err error) {
	if err = d.pingMC(ctx); err != nil {
		return
	}
	if err = d.pingRedis(ctx); err != nil {
		return
	}
	return d.db.Ping(ctx)
}

func (d *Dao) pingMC(ctx context.Context) (err error) {
	if err = d.mc.Set(ctx, &memcache.Item{Key: "ping", Value: []byte("pong"), Expiration: 0}); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}

func (d *Dao) pingRedis(ctx context.Context) (err error) {
	conn := d.redis.Get(ctx)
	defer conn.Close()
	if _, err = conn.Do("SET", "ping", "pong"); err != nil {
		log.Error("conn.Set(PING) error(%v)", err)
	}
	return
}
`
	_tplReadme = `# {{.Name}}

## 项目简介
1.
`
	_tplService = `package service

import (
	"context"

	"{{.Name}}/internal/dao"
	"github.com/bilibili/kratos/pkg/conf/paladin"
)

// Service service.
type Service struct {
	ac  *paladin.Map
	dao *dao.Dao
}

// New new a service and return.
func New() (s *Service) {
	var ac = new(paladin.TOML)
	if err := paladin.Watch("application.toml", ac); err != nil {
		panic(err)
	}
	s = &Service{
		ac:  ac,
		dao: dao.New(),
	}
	return s
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}
`

	_tplGPRCService = `package service

import (
	"context"
	"fmt"

	pb "{{.Name}}/api"
	"{{.Name}}/internal/dao"
	"github.com/bilibili/kratos/pkg/conf/paladin"

	"github.com/golang/protobuf/ptypes/empty"
)

// Service service.
type Service struct {
	ac  *paladin.Map
	dao *dao.Dao
}

// New new a service and return.
func New() (s *Service) {
	var ac = new(paladin.TOML)
	if err := paladin.Watch("application.toml", ac); err != nil {
		panic(err)
	}
	s = &Service{
		ac:  ac,
		dao: dao.New(),
	}
	return s
}

// SayHello grpc demo func.
func (s *Service) SayHello(ctx context.Context, req *pb.HelloReq) (reply *empty.Empty, err error) {
	reply = new(empty.Empty)
	fmt.Printf("hello %s", req.Name)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context) (err error) {
	return s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
	s.dao.Close()
}
`
	_tplHTTPServer = `package http

import (
	"net/http"

	"{{.Name}}/internal/service"
	"github.com/bilibili/kratos/pkg/conf/paladin"
	"github.com/bilibili/kratos/pkg/log"
	bm "github.com/bilibili/kratos/pkg/net/http/blademaster"
	"github.com/bilibili/kratos/pkg/net/http/blademaster/middleware/verify"
)

var (
	svc *service.Service
)

// New new a bm server.
func New(s *service.Service) (engine *bm.Engine) {
	var (
		hc struct {
			Server *bm.ServerConfig
		}
	)
	if err := paladin.Get("http.toml").UnmarshalTOML(&hc); err != nil {
		if err != paladin.ErrNotExist {
			panic(err)
		}
	}
	svc = s
	engine = bm.DefaultServer(hc.Server)
	initRouter(engine, verify.New(nil))
	if err := engine.Start(); err != nil {
		panic(err)
	}
	return
}

func initRouter(e *bm.Engine, v *verify.Verify) {
	e.Ping(ping)
	e.Register(register)
	g := e.Group("/{{.Name}}")
	{
		g.GET("/start", v.Verify, howToStart)
	}
}

func ping(ctx *bm.Context) {
	if err := svc.Ping(ctx); err != nil {
		log.Error("ping error(%v)", err)
		ctx.AbortWithStatus(http.StatusServiceUnavailable)
	}
}

func register(c *bm.Context) {
	c.JSON(map[string]interface{}{}, nil)
}

// example for http request handler.
func howToStart(c *bm.Context) {
	c.String(0, "Golang 大法好 !!!")
}
`
	_tplAPIProto = `// 定义项目 API 的 proto 文件 可以同时描述 gRPC 和 HTTP API
// protobuf 文件参考:
//  - https://developers.google.com/protocol-buffers/
//  - TODO：待补充文档URL
// protobuf 生成 HTTP 工具:
//  - TODO：待补充文档URL
// gRPC Golang Model:
//  - TODO：待补充文档URL
// gRPC Golang Warden Gen:
//  - TODO：待补充文档URL
// gRPC http 调试工具(无需pb文件):
//  - TODO：待补充文档URL
// grpc 命令行调试工具(无需pb文件):
//  - TODO：待补充文档URL
syntax = "proto3";

import "github.com/gogo/protobuf/gogoproto/gogo.proto";
import "google/protobuf/empty.proto";

// package 命名使用 {appid}.{version} 的方式, version 形如 v1, v2 ..
package demo.service.v1;

// NOTE: 最后请删除这些无用的注释 (゜-゜)つロ 

option go_package = "api";
// do not generate getXXX() method 
option (gogoproto.goproto_getters_all) = false;

service Demo {
	rpc SayHello (HelloReq) returns (.google.protobuf.Empty);
}

message HelloReq {
	string name = 1 [(gogoproto.moretags)='form:"name" validate:"required"'];
}
`
	_tplAPIGenerate = `package api

// 生成 gRPC 代码
//go:generate TODO:待完善工具protoc.sh
`
	_tplModel = `package model
`
	_tplGRPCServer = `package grpc

import (
	pb "{{.Name}}/api"
	"{{.Name}}/internal/service"
	"github.com/bilibili/kratos/pkg/conf/paladin"
	"github.com/bilibili/kratos/pkg/net/rpc/warden"
)

// New new a grpc server.
func New(svc *service.Service) *warden.Server {
	var rc struct {
		Server *warden.ServerConfig
	}
	if err := paladin.Get("grpc.toml").UnmarshalTOML(&rc); err != nil {
		if err != paladin.ErrNotExist {
			panic(err)
		}
	}
	ws := warden.NewServer(rc.Server)
	pb.RegisterDemoServer(ws.Server(), svc)
	ws, err := ws.Start()
	if err != nil {
		panic(err)
	}
	return ws
}
`
)
