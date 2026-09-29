package main

// @title Todo API
// @version 1.0
// @description Todo 项目 REST API
// @host localhost:8080
// @BasePath /

import (
	"bufio"
	"fmt"
	"os"
	"todo/service"
	"todo/storage"
	"todo/web"
)

func main() {
	storage.InitDB()
	web.StartServer()
	reader := bufio.NewReader(os.Stdin)
	todos := storage.LoadTodo()
	for {
		fmt.Println("================")
		fmt.Println("    Todo List")
		fmt.Println("================")
		fmt.Println("1. 添加任务")
		fmt.Println("2. 查看任务")
		fmt.Println("3. 完成任务")
		fmt.Println("4. 删除任务")
		fmt.Println("5. 修改任务")
		fmt.Println("6. 搜索任务")
		fmt.Println("7. 查看统计")
		fmt.Println("8. 退出")
		fmt.Print("请选择：")
		input, _ := reader.ReadString('\n')
		var choice int
		fmt.Sscanf(input, "%d", &choice)
		switch choice {
		case 1:
			todos = service.AddTodo(todos, reader)
		case 2:
			service.ListTodo(todos)
		case 3:
			service.CompleteTodo(todos, reader)
		case 4:
			todos = service.DeleteTodo(todos, reader)
		case 5:
			todos = service.UpdateTodo(todos, reader)
		case 6:
			service.SearchTodoCLI(todos, reader)
		case 7:
			service.StaTodo(todos)
		case 8:
			fmt.Println("退出程序")
			return
		default:
			fmt.Println("请输入 1-3")
		}
		fmt.Println()
	}
}
