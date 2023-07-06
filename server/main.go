package server

import (
	"github.com/iris-contrib/middleware/cors"
	iris "github.com/kataras/iris/v12"
)

type WebServer struct {
	Port int32
}

type Process struct {
	Name        string  `json:"name"`
	Pid         int32   `json:"pid"`
	Status      string  `json:"status"`
	CpuUsage    float32 `json:"cpuUsage"`
	MemoryUsage float32 `json:"memoryUsage"`
}

func handleApiRequestForProcessList(ctx iris.Context) {
	p := Process{
		Name:        "test",
		Pid:         123,
		Status:      "running",
		CpuUsage:    0.1,
		MemoryUsage: 0.2,
	}

	ctx.JSON(p)
}

func handleApiRequestForScheduledTasks(ctx iris.Context) {
	p := Process{
		Name:        "task N",
		Pid:         123,
		Status:      "running every 2 minutes",
		CpuUsage:    0.1,
		MemoryUsage: 0.2,
	}

	tasks := make([]Process, 0)
	tasks = append(tasks, p)
	tasks = append(tasks, p)

	ctx.JSON(tasks)
}

func (s *WebServer) Start() {
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	app := iris.New()
	app.Use(iris.Compression)
	app.UseRouter(crs)
	app.HandleDir("/", iris.Dir("./dist/web"))
	app.Get("/api/process/status", handleApiRequestForProcessList)
	app.Get("/api/scheduled-tasks", handleApiRequestForScheduledTasks)
	go app.Listen(":8500")
}
