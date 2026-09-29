package model

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title" binding:"required,min=1,max=100"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
	Priority  int    `json:"priority" binding:"min=1,max=5"`
	DueDate    string `json:"due_date" binding:"omitempty,datetime=2006-01-02"`
}
type BatchCompleteRequest struct {
	IDs []int `json:"ids"`
}