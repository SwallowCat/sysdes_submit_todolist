package service

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	database "todolist.go/db"
)
 
func NewUserForm(ctx *gin.Context) {
    ctx.HTML(http.StatusOK, "new_user_form.html", gin.H{"Title": "Register user"})
}

func hash(pw string) []byte {
    const salt = "todolist.go#"
    h := sha256.New()
    h.Write([]byte(salt))
    h.Write([]byte(pw))
    return h.Sum(nil)
}

func RegisterUser(ctx *gin.Context) {
	//フォームデータの受け取り
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	passwordConfirm := ctx.PostForm("password_confirm")
	switch {
    case username == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Usernane is not provided", "Username": username})
    case password == "":
        ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is not provided", "Password": password})
	case password != passwordConfirm:
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password does not match", "Username": username, "Password": password})
		return
    }

	//パスワードの長さチェック
	if len(password) < 5 {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password is too short", "Username": username, "Password": password})
		return
	}
	//パスワードが数字とアルファベットを含むかチェック
	var hasNumber bool
	var hasAlphabet bool
	for _, c := range password {
		switch {
		case '0' <= c && c <= '9':
			hasNumber = true
		case 'a' <= c && c <= 'z':
			hasAlphabet = true
		}
	}
	if !hasNumber || !hasAlphabet {
		ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Password must contain at least one number and one alphabet", "Username": username, "Password": password})
		return
	}

	//DB 接続
	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	 // 重複チェック
	 var duplicate int
	 err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=?", username)
	 if err != nil {
		 Error(http.StatusInternalServerError, err.Error())(ctx)
		 return
	 }
	 if duplicate > 0 {
		 ctx.HTML(http.StatusBadRequest, "new_user_form.html", gin.H{"Title": "Register user", "Error": "Username is already taken", "Username": username, "Password": password})
		 return
	 }

	// DB への保存
	result, err := db.Exec("INSERT INTO users (name, password) VALUES (?, ?)", username, hash(password))
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	//保存状態の確認
	id , _ := result.LastInsertId()
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", id)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	//ctx.JSON(http.StatusOK, user)

	//ログイン画面へリダイレクト
	ctx.Redirect(http.StatusFound, "/login")
}

func LoginForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})
}

const userkey = "user"
 
func Login(ctx *gin.Context) {
    username := ctx.PostForm("username")
    password := ctx.PostForm("password")
 
    db, err := database.GetConnection()
    if err != nil {
        Error(http.StatusInternalServerError, err.Error())(ctx)
        return
    }
 
    // ユーザの取得
    var user database.User
    err = db.Get(&user, "SELECT id, name, password FROM users WHERE name = ? AND is_deleted=0", username)
    if err != nil {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "No such user"})
        return
    }
 
    // パスワードの照合
    if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
        ctx.HTML(http.StatusBadRequest, "login.html", gin.H{"Title": "Login", "Username": username, "Error": "Incorrect password"})
        return
    }
 
    // セッションの保存
    session := sessions.Default(ctx)
    session.Set(userkey, user.ID)
    session.Save()
 
    ctx.Redirect(http.StatusFound, "/list")
}

func LoginCheck(ctx *gin.Context) {
    if sessions.Default(ctx).Get(userkey) == nil {
        ctx.Redirect(http.StatusFound, "/login")
        ctx.Abort()
    } else {
        ctx.Next()
    }
}

func Logout(ctx *gin.Context) {
    session := sessions.Default(ctx)
    session.Clear()
    session.Options(sessions.Options{MaxAge: -1})
    session.Save()
    ctx.Redirect(http.StatusFound, "/")
}

func EditUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "edit_user_form.html", gin.H{"Title": "Edit user"})
}

func EditUserNameForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "edit_user_name.html", gin.H{"Title": "Edit user name"})
}

func EditUserName(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	username := ctx.PostForm("name")
	password := ctx.PostForm("password")

	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_name.html", gin.H{"Title": "Edit user name", "Error": "Usernane is not provided", "Username": username})
	case password == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_name.html", gin.H{"Title": "Edit user name", "Error": "Password is not provided", "Password": password})
	}

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ? AND is_deleted=0", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "edit_user_name.html", gin.H{"Title": "Edit user name", "Username": username, "Error": "Incorrect password"})
		return
	}

	// 重複チェック
	var duplicate int
	err = db.Get(&duplicate, "SELECT COUNT(*) FROM users WHERE name=? AND is_deleted=0", username)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	if duplicate > 0 {
		ctx.HTML(http.StatusBadRequest, "edit_user_name.html", gin.H{"Title": "Edit user name", "Error": "Username is already taken", "Username": username})
		return
	}

	// DB への保存
	_, err = db.Exec("UPDATE users SET name = ? WHERE id = ?", username, userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.Redirect(http.StatusFound, "/list")
	/* err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}
	ctx.JSON(http.StatusOK, user) */
}

func EditUserPasswordForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "edit_user_password.html", gin.H{"Title": "Edit user password"})
}

func EditUserPassword(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	username := ctx.PostForm("name")
	old_password := ctx.PostForm("old_password")
	new_password := ctx.PostForm("new_password")

	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_password.html", gin.H{"Title": "Edit user password", "Error": "Usernane is not provided", "Username": username})
	case old_password == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_password.html", gin.H{"Title": "Edit user password", "Error": "Password is not provided", "Password": old_password})
	case new_password == "":
		ctx.HTML(http.StatusBadRequest, "edit_user_password.html", gin.H{"Title": "Edit user password", "Error": "Password is not provided", "Password": new_password})
	}

	//new_passwordの長さチェック
	if len(new_password) < 5 {
		ctx.HTML(http.StatusBadRequest, "edit_user_password.html", gin.H{"Title": "Edit user password", "Error": "Password is too short", "Username": username, "Password": new_password})
		return
	}
	//new_passwordが数字とアルファベットを含むかチェック
	var hasNumber bool
	var hasAlphabet bool
	for _, c := range new_password {
		switch {
		case '0' <= c && c <= '9':
			hasNumber = true
		case 'a' <= c && c <= 'z':
			hasAlphabet = true
		}
	}
	if !hasNumber || !hasAlphabet {
		ctx.HTML(http.StatusBadRequest, "edit_user_password.html", gin.H{"Title": "Edit user password", "Error": "Password must contain at least one number and one alphabet", "Username": username, "Password": new_password})
		return
	}

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ? AND is_deleted=0", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(old_password)) {
		ctx.HTML(http.StatusBadRequest, "edit_user_password.html", gin.H{"Title": "Edit user password", "Username": username, "Error": "Incorrect password"})
		return
	}

	// DB への保存
	_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hash(new_password), userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.Redirect(http.StatusFound, "/list")
}

func DeleteUserForm(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "delete_user_form.html", gin.H{"Title": "Delete user"})
}

func DeleteUser(ctx *gin.Context) {
	userID := sessions.Default(ctx).Get("user")
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	switch {
	case username == "":
		ctx.HTML(http.StatusBadRequest, "delete_user_form.html", gin.H{"Title": "Delete user", "Error": "Usernane is not provided", "Username": username})
	case password == "":
		ctx.HTML(http.StatusBadRequest, "delete_user_form.html", gin.H{"Title": "Delete user", "Error": "Password is not provided", "Password": password})
	}

	db, err := database.GetConnection()
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// ユーザの取得
	var user database.User
	err = db.Get(&user, "SELECT id, name, password FROM users WHERE id = ? AND is_deleted=0", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	// パスワードの照合
	if hex.EncodeToString(user.Password) != hex.EncodeToString(hash(password)) {
		ctx.HTML(http.StatusBadRequest, "delete_user_form.html", gin.H{"Title": "Delete user", "Username": username, "Error": "Incorrect password"})
		return
	}

	// DB への保存
	_, err = db.Exec("UPDATE users SET is_deleted = 1 WHERE id = ?", userID)
	if err != nil {
		Error(http.StatusInternalServerError, err.Error())(ctx)
		return
	}

	ctx.Redirect(http.StatusFound, "/logout")
}