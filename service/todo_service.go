package service

import (
	"bufio"
	"fmt"
	"sort"
	"strings"
	"time"
	"todo/model"
	"todo/storage"
	"todo/utils"
)

func AddTodo(todos []model.Todo, reader *bufio.Reader) []model.Todo {
	fmt.Print("请输入任务：")
	title, _ := reader.ReadString('\n')
	title = strings.TrimSpace(title)
	fmt.Print("请输入优先级(1高 2中 3低):")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	var priority int
	fmt.Sscanf(input, "%d", &priority)
	todo := model.Todo{
		ID:        utils.GetNextID(todos),
		Title:     title,
		Done:      false,
		CreatedAt: time.Now().Format("2006-01-02 15:04"),
		Priority:  priority,
	}
	todos = append(todos, todo)
	storage.CreateTodo(todo)
	fmt.Println("添加成功！")
	return todos
}

func ListTodo(todos []model.Todo) {
	sortedTodos := make([]model.Todo, len(todos))
	copy(sortedTodos, todos)
	sort.Slice(sortedTodos, func(i, j int) bool {
		return sortedTodos[i].Priority < sortedTodos[j].Priority
	})
	fmt.Println("当前任务：")
	if len(todos) == 0 {
		fmt.Println("暂无任务")
	} else {
		for _, todo := range sortedTodos {
			status := "[ ]"
			if todo.Done {
				status = "[✓]"
			}
			fmt.Println(todo.ID, status, todo.Title)
			fmt.Println("   优先级:", utils.PriorityText(todo.Priority))
			fmt.Println("   创建时间:", todo.CreatedAt)
		}
	}
}
func CompleteTodo(todos []model.Todo, reader *bufio.Reader) {
	fmt.Println("当前任务：")
	for i, todo := range todos {
		fmt.Println(i+1, todo.Title)
	}
	fmt.Println("请输入任务ID:")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	var id int
	fmt.Sscanf(input, "%d", &id)
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Done = true
			storage.UpdateTodo(todos[i])
			fmt.Println("任务已完成！")
			return
		}
	}
}
func DeleteTodo(todos []model.Todo, reader *bufio.Reader) []model.Todo {
	fmt.Println("当前任务：")
	for i, todo := range todos {
		fmt.Println(i+1, todo.Title)
	}
	fmt.Println("请输入删除任务编号：")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	var index int
	fmt.Sscanf(input, "%d", &index)
	for i, todo := range todos {
		if todo.ID == index {
			storage.DeleteTodo(index)
			todos = append(todos[:i], todos[i+1:]...)
			fmt.Println("删除成功！")
			return todos
		}
	}
	fmt.Println("没有找到该任务")
	return todos
}

func UpdateTodo(todos []model.Todo, reader *bufio.Reader) []model.Todo {
	fmt.Print("请输入任务ID:")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	var id int
	fmt.Sscanf(input, "%d", &id)
	fmt.Println("读取到的ID:", id)
	for i := range todos {
		if todos[i].ID == id {
			fmt.Print("请输入新的任务内容:")
			title, _ := reader.ReadString('\n')
			todos[i].Title = title
			storage.UpdateTodo(todos[i])
			fmt.Println("修改成功！")
			return todos
		}
	}
	fmt.Println("没有找到该任务")
	return todos
}
func SearchTodoCLI(todos []model.Todo, reader *bufio.Reader) {
	fmt.Print("请输入搜索关键词:")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)
	fmt.Println("搜索结果:")
	for _, todo := range todos {
		if strings.Contains(todo.Title, keyword) {
			status := "[ ]"
			if todo.Done {
				status = "[✓]"
			}
			fmt.Println(todo.ID, status, todo.Title)
			fmt.Println("    创建时间:", todo.CreatedAt)
		}
	}
}

func StaTodo(todos []model.Todo) {
	total := len(todos)
	done := 0
	high := 0
	middle := 0
	low := 0
	for _, todo := range todos {
		if todo.Done {
			done++
		}
		switch todo.Priority {
		case 1:
			high++
		case 2:
			middle++
		case 3:
			low++
		}
	}
	fmt.Println("======任务统计======")
	fmt.Println("总任务:", total)
	fmt.Println("已完成:", done)
	fmt.Println("未完成:", total-done)
	fmt.Println()
	fmt.Println("高优先级:", high)
	fmt.Println("中优先级:", middle)
	fmt.Println("低优先级:", low)
}
func CreateTodo(todo model.Todo) bool {
	return storage.CreateTodo(todo)
}
func GetTodos() []model.Todo {
	return storage.LoadTodo()
}
func GetTodoByID(id int) (model.Todo, bool) {
	return storage.GetTodoByID(id)
}
func UpdateTodoByID(todo model.Todo) bool {
	return storage.UpdateTodo(todo)
}
func DeleteTodoByID(id int) bool {
	return storage.DeleteTodo(id)
}
func CompleteTodoByID(id int) bool {
	return storage.CompleteTodo(id)
}
func QueryTodos(keyword string, page int, pageSize int, sort string, done *bool) []model.Todo {
	return storage.QueryTodos(keyword, page, pageSize, sort, done)
}
func CountTodos(keyword string) int {
	return storage.CountTodos(keyword)
}
func BatchCompleteTodos(ids []int) error {
	err := storage.BatchCompleteTodos(ids)
	if err != nil {
		return utils.NewAppError(404, err.Error())
	}
	return nil
}