package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
    store := cookie.NewStore([]byte("my-secret"))
    engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	
	taskGroup := engine.Group("/task")
    taskGroup.Use(service.LoginCheck)
    {
        taskGroup.GET("/:id", service.ShowTask)
        taskGroup.GET("/new", service.NewTaskForm)
        taskGroup.POST("/new", service.RegisterTask)
        taskGroup.GET("/edit/:id", service.EditTaskForm)
        taskGroup.POST("/edit/:id", service.UpdateTask)
        taskGroup.GET("/delete/:id", service.DeleteTask)
    }

	// ユーザ登録
    engine.GET("/user/new", service.NewUserForm)
    engine.POST("/user/new", service.RegisterUser)

	// ユーザー情報変更
	engine.GET("/user/edit", service.LoginCheck, service.EditUserForm)

	//ユーザー名変更
	engine.GET("/user/edit/name", service.EditUserNameForm)
	engine.POST("/user/edit/name", service.EditUserName)

	//パスワード変更
	engine.GET("/user/edit/password", service.EditUserPasswordForm)
	engine.POST("/user/edit/password", service.EditUserPassword)

	// ログイン
	engine.GET("/login", service.LoginForm)
	engine.POST("/login", service.Login)

	// ログアウト
	engine.GET("/logout", service.Logout)

	//　ユーザー削除
	engine.GET("/user/delete", service.DeleteUserForm)
	engine.POST("/user/delete", service.DeleteUser)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
