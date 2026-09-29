package utils

import (
	"todo/model"
)

func GetNextID(todos []model.Todo) int {
	maxID := 0
	for _, todo := range todos {
		if todo.ID > maxID {
			maxID = todo.ID
		}
	}
	return maxID + 1
}
func PriorityText(priority int) string {
	switch priority {
	case 1:
		return "高"
	case 2:
		return "中"
	case 3:
		return "低"
	default:
		return "未知"
	}
}
