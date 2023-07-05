package server

import (
	iris "github.com/kataras/iris/v12"
)

type WebServer struct {
	Port int32
}

func (s *WebServer) Start() {
	app := iris.New()
	app.Use(iris.Compression)
	app.HandleDir("/", iris.Dir("./dist/web"))
	go app.Listen(":8500")
}
