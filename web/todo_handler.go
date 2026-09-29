package web

import (
	"todo/model"
	"strconv"
	"todo/service"
	"github.com/gin-gonic/gin"
	"todo/utils"
	"strings"
)

// GetTodos
// @Summary 查询Todo
// @Description 支持关键词搜索、分页、排序和完成状态筛选
// @Tags Todo
// @Produce json
// @Param keyword query string false "搜索关键词"
// @Param page query int false "页码" 
// @Param pageSize query int false "每页数量"
// @Param sort query string false "排序方式"
// @Param done query bool false "完成状态"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /todos [get]
func GetTodos(c *gin.Context) {
	keyword:=c.DefaultQuery("keyword","")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "10")
	sort:=c.DefaultQuery("sort","priority_desc")
	doneStr := c.Query("done")
	var done *bool
	if doneStr != "" {
		value, err := strconv.ParseBool(doneStr)
		if err != nil {
			utils.Error(c, 400, "done参数错误")
			return
		}
		done = &value
	}
	if sort != "priority_desc" && 
		sort != "priority_asc" &&
		sort != "due_date_desc" &&
		sort != "due_date_asc" {
		utils.Error(c, 400, "sort参数错误")
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		utils.Error(c, 400, "page格式错误")
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		utils.Error(c, 400, "pageSize必须为1-100")
		return
	}
	todos := service.QueryTodos(keyword, page, pageSize, sort, done)
	total := service.CountTodos(keyword)
	totalPages := (total + pageSize - 1) / pageSize
	utils.Success(c, gin.H{
		"todos":      todos,
		"page":       page,
		"page_size":   pageSize,
		"total":      total,
		"total_pages": totalPages,
	})
}
func GetTodoByID(c *gin.Context) {
	id := c.Param("id")
	todoID, err := strconv.Atoi(id)
	if err != nil {
		utils.Error(c, 400, "ID格式错误")
		return
	}
	todo, success := service.GetTodoByID(todoID)
	if !success {
		utils.Error(c, 404, "任务不存在")
		return
	}
	utils.Success(c, todo)
}
// CreateTodo
// @Summary 创建 Todo
// @Description 创建一个新的 Todo
// @Tags Todo
// @Accept json
// @Produce json
// @Param todo body model.Todo true "Todo"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /todos [post]
func CreateTodo(c *gin.Context) {
	var todo model.Todo
	err := c.ShouldBindJSON(&todo)
	if err != nil {
		utils.Error(c, 400, "参数校验失败")
		return
	}
	if strings.TrimSpace(todo.Title) == "" {
		utils.Error(c, 400, "任务标题不能为空")
		return
	}
	if todo.Priority < 1 || todo.Priority > 3 {
		utils.Error(c, 400, "优先级必须为1、2或3")
		return
	}
	success := service.CreateTodo(todo)
	if !success {
		utils.Error(c, 500, "创建任务失败")
		return
	}
	utils.Success(c, gin.H{
		"message": "添加成功",
	})
}
// UpdateTodo
// @Summary 修改 Todo
// @Description 修改指定 Todo
// @Tags Todo
// @Accept json
// @Produce json
// @Param id path int true "Todo ID"
// @Param todo body model.Todo true "Todo"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /todos/{id} [put]
func UpdateTodo(c *gin.Context) {
	id:= c.Param("id")
	var todo model.Todo
	err := c.ShouldBindJSON(&todo)
	if err != nil {
		utils.Error(c, 400, "参数校验失败")
		return
		}
		todo.ID,err= strconv.Atoi(id)
		if err != nil {
			utils.Error(c, 400, "ID格式错误")
			return
		}
		success := service.UpdateTodoByID(todo)
		if !success {
			utils.Error(c, 404, "任务不存在")
			return
		}
		if strings.TrimSpace(todo.Title) == "" {
			utils.Error(c, 400, "任务标题不能为空")
			return
		}
		if todo.Priority < 1 || todo.Priority > 3 {
			utils.Error(c, 400, "优先级必须为1、2或3")
			return
		}
		utils.Success(c, gin.H{
			"message": "修改成功",
		})
}
// DeleteTodo
// @Summary 删除 Todo
// @Description 删除指定 Todo
// @Tags Todo
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /todos/{id} [delete]
func DeleteTodo(c *gin.Context) {
	id:= c.Param("id")
	todoID,err := strconv.Atoi(id)
	if err != nil {
		utils.Error(c, 400, "ID格式错误")
		return
	}
	success := service.DeleteTodoByID(todoID)
	if !success {
		utils.Error(c, 404, "任务不存在")
		return
	}
	utils.Success(c, gin.H{
		"message": "删除成功",
	})
}
// CompleteTodo
// @Summary 完成 Todo
// @Description 将指定 Todo 标记为完成
// @Tags Todo
// @Produce json
// @Param id path int true "Todo ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /todos/{id}/complete [put]
func CompleteTodo(c *gin.Context) {
	id:= c.Param("id")
	todoID,err := strconv.Atoi(id)
	if err != nil {
		utils.Error(c, 400, "ID格式错误")
		return
	}
	success := service.CompleteTodoByID(todoID)
	if !success {
		utils.Error(c, 404, "任务不存在")
		return
	}
	utils.Success(c, gin.H{
		"message": "完成成功",
	})
}
func BatchCompleteTodos(c *gin.Context) {
	var req model.BatchCompleteRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		utils.Error(c, 400, "参数错误")
		return
	}
	if len(req.IDs) == 0 {
		utils.Error(c, 400, "任务ID不能为空")
		return
	}
	err = service.BatchCompleteTodos(req.IDs)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			utils.Error(c, appErr.Code, appErr.Message)
			return
		}
		utils.Error(c,500,"服务器内部错误")
		return
	}
	utils.Success(c, gin.H{
		"message": "批量完成成功",
	})
}