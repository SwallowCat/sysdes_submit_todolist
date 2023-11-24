package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)

// TaskList renders list of tasks in DB
func TaskList(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	//Get query parameter
	kw := ctx.Query("kw")

	is_done, exist := ctx.GetQuery("is_done")

	query := "SELECT id, title, created_at, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
	
	if exist{
		if is_done == "all"{
			// Get tasks in DB
			var tasks []database.Task
			switch {
			case kw != "":
				//err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ?", "%"+kw+"%")
				err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
			default:
				//err = db.Select(&tasks, "SELECT * FROM tasks")
				err = db.Select(&tasks, query, userID)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}

			ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks})
		}else{
			is_done_bool, err := strconv.ParseBool(is_done)
			if err != nil{
				Error(http.StatusBadRequest, err.Error())(ctx)
				return
			}
			//Get tasks in DB
			var tasks []database.Task
			switch{
			case kw != "":
				//err = db.Select(&tasks, "SELECT * FROM tasks WHERE title LIKE ? AND is_done=?", "%"+kw+"%", is_done_bool)
				err = db.Select(&tasks, query + " AND title LIKE ?" + " AND is_done=?", userID, "%" + kw + "%", is_done_bool)
			default:
				//err = db.Select(&tasks, "SELECT * FROM tasks WHERE is_done=?", is_done_bool)
				err = db.Select(&tasks, query + " AND is_done=?", userID, is_done_bool)
			}
			if err != nil {
				Error(http.StatusInternalServerError, err.Error())(ctx)
				return
			}

			ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks})
		}
	}else{
		// Get tasks in DB
		var tasks []database.Task
		switch {
		case kw != "":
			err = db.Select(&tasks, query + " AND title LIKE ?", userID, "%" + kw + "%")
		default:
			err = db.Select(&tasks, query, userID)
		}
		if err != nil {
			Error(http.StatusInternalServerError, err.Error())(ctx)
			return
		}

		ctx.HTML(http.StatusOK, "task_list.html", gin.H{"Title": "Task list", "Tasks": tasks})
	}
}

// ShowTask renders a task with given ID
func ShowTask(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	query := "SELECT id, title, created_at, is_done FROM tasks INNER JOIN ownership ON task_id = id WHERE user_id = ?"
	// Get DB connection
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// parse ID given as a parameter
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Get a task with given ID
	var task database.Task
	//err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id) // Use DB#Get for one entry
	err = db.Get(&task, query + " AND id=?", userID, id)
	if err != nil {
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}

	// Render task
	//ctx.String(http.StatusOK, task.Title)  // Modify it!!
	ctx.HTML(http.StatusOK, "task.html", task)
}

func NewTaskForm(ctx *gin.Context){
	ctx.HTML(http.StatusOK, "form_new_task.html", gin.H{"Title": "Task registeration"})
}

func RegisterTask(ctx *gin.Context){
	userID := sessions.Default(ctx).Get("user")
	//Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist{
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	tx := db.MustBegin()
	//Create new data with given title on DB
	result, err := db.Exec("INSERT INTO tasks (title) VALUES (?)", title)
	if err != nil{
		tx.Rollback()
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	taskID, err := result.LastInsertId()
	if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
	_, err = tx.Exec("INSERT INTO ownership (user_id, task_id) VALUES (?, ?)", userID, taskID)
    if err != nil {
        tx.Rollback()
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
    tx.Commit()
    ctx.Redirect(http.StatusFound, fmt.Sprintf("/task/%d", taskID))
}

func EditTaskForm(ctx *gin.Context){
	//IDの取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//get DB connection
	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Get target task
	var task database.Task
	err = db.Get(&task, "SELECT * FROM tasks WHERE id=?", id)
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Render edit form
	ctx.HTML(http.StatusOK, "form_edit_task.html", 
		gin.H{"Title": fmt.Sprintf("Edit task %d", task.ID), "Task": task})
}

func UpdateTask(ctx *gin.Context){
	//Get Id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get task title
	title, exist := ctx.GetPostForm("title")
	if !exist{
		Error(http.StatusBadRequest, "No title is given")(ctx)
		return
	}
	//Get task is_done
	is_done, exist := ctx.GetPostForm("is_done")
	if !exist{
		Error(http.StatusBadRequest, "No is_done is given")(ctx)
		return
	}
	is_done_bool, err := strconv.ParseBool(is_done)
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Update task
	result, err := db.Exec("UPDATE tasks SET title=?, is_done=? WHERE id=?", title, is_done_bool, id)
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Render status
	path := "/list"
	if rows, _ := result.RowsAffected(); rows == 1{
		path = fmt.Sprintf("/task/%d", id)
	}
	ctx.Redirect(http.StatusFound, path)
}

func DeleteTask(ctx *gin.Context){
	//idの取得
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil{
		Error(http.StatusBadRequest, err.Error())(ctx)
		return
	}
	//Get DB connection
	db, err := database.GetConnection()
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Delete the task from DB
	_, err = db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil{
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//Render status
	ctx.Redirect(http.StatusFound, "/list")
}
