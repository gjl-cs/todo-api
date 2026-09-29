package storage

import (
	"todo/model"
	"fmt"
	"strings"
)

func LoadTodo() []model.Todo {
	rows, err := DB.Query(
		"SELECT id,title,done,priority,created_at,COALESCE(due_date, '') FROM todos",
	)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var todos []model.Todo
	for rows.Next() {
		var todo model.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Done,
			&todo.Priority,
			&todo.CreatedAt,
			&todo.DueDate,
		)
		todo.Title=strings.TrimSpace(todo.Title)
		if err != nil {
			panic(err)
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}
	return todos
}
func GetTodoByID(id int) (model.Todo, bool) {
	var todo model.Todo
	err := DB.QueryRow(
		"SELECT id,title,done,priority,created_at,COALESCE(due_date, '') FROM todos WHERE id=?",
		id,
	).Scan(
		&todo.ID,
		&todo.Title,
		&todo.Done,
		&todo.Priority,
		&todo.CreatedAt,
		&todo.DueDate,
	)
	if err!=nil{
		return todo,false
	}
	todo.Title=strings.TrimSpace(todo.Title)
	return todo,true
}
func CreateTodo(todo model.Todo) bool {
	sql := `
	INSERT INTO todos
	(title,done,priority,created_at,due_date)
	VALUES(?,?,?,?,?)
	`
	_, err := DB.Exec(
		sql,
		todo.Title,
		todo.Done,
		todo.Priority,
		todo.CreatedAt,
		todo.DueDate,
	)
	if err != nil {
		fmt.Println("创建任务失败:", err)
		return false
	}
	return true
}
func UpdateTodo(todo model.Todo) bool{
	result, err := DB.Exec(
		"UPDATE todos SET title=?,done=?,priority=? WHERE id=?",
		todo.Title,
		todo.Done,
		todo.Priority,
		todo.DueDate,
		todo.ID,
	)
	if err != nil {
		return false
	}
	rows,err:=result.RowsAffected()
	if err != nil {
		return false
	}
	return rows > 0
}
func DeleteTodo(id int) bool {
	result, err := DB.Exec(
		"DELETE FROM todos WHERE id=?",
		id,
	)
	if err != nil {
		fmt.Println("删除任务失败:", err)
		return false
	}
	rows,_ := result.RowsAffected()
	return rows > 0
}
func CompleteTodo(id int) bool {
	result, err := DB.Exec(
		"UPDATE todos SET done = 1 WHERE id = ?",
		id,
	)
	if err != nil {
		fmt.Println("完成任务失败:", err)
		return false
	}
	rows, _ := result.RowsAffected()
	return rows > 0
}
func QueryTodos(keyword string, page int, pageSize int, sort string, done *bool) []model.Todo {
	offset := (page - 1) * pageSize
	orderBy:="priority DESC"
	switch sort {
	case "priority_asc":
		orderBy = "priority ASC"
	case "priority_desc":
		orderBy = "priority DESC"
	case "due_date_asc":
		orderBy = "due_date = '' ASC,due_date ASC"
	case "due_date_desc":
		orderBy = "due_date = '' ASC,due_date DESC"
	}
	query := `
	SELECT id,title,done,priority,created_at,COALESCE(due_date, '') 
	FROM todos 
	WHERE title LIKE ?
	`
	args := []interface{}{
		"%" + keyword + "%",
	}
	if done != nil {
		query += " AND done = ?"
		args = append(args, *done)
	}
	query += " ORDER BY " + orderBy + " LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		fmt.Println("查询Todo失败:", err)
		return nil
	}
	defer rows.Close()
	todos := make([]model.Todo, 0)
	for rows.Next() {
		var todo model.Todo
		err := rows.Scan(
			&todo.ID,
			&todo.Title,
			&todo.Done,
			&todo.Priority,
			&todo.CreatedAt,
			&todo.DueDate,
		)
		if err != nil {
			fmt.Println("查询Todo失败:", err)
			return nil
		}
		todo.Title = strings.TrimSpace(todo.Title)
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		fmt.Println("查询结果错误:", err)
		return nil
	}
	return todos
}
func CountTodos(keyword string) int {
	var total int
	err := DB.QueryRow(
		"SELECT COUNT(*) FROM todos WHERE title LIKE ?",
		"%"+keyword+"%",
	).Scan(&total)
	if err != nil {
		return 0
	}
	return total
}
func BatchCompleteTodos(ids []int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	for _, id := range ids {
		result, err := tx.Exec(
			"UPDATE todos SET done = 1 WHERE id = ?",
			id,
		)
		if err != nil {
			tx.Rollback()
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			return err
		}
		if rows == 0 {
			tx.Rollback()
			return fmt.Errorf("任务ID %d 不存在", id)
		}
	}
	return tx.Commit()
}