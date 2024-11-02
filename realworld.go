package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type ILogindata interface {
}

type FakeTime struct {
	Valid bool
}

type ISessionManager interface {
	Create(*http.Request) (*Session, error)
	Check(*http.Request, *User) error
}

type Session struct {
	SessionID int
}

type SessionManager struct {
	Sessions []Session
}

type LoginData struct {
	Login    string
	Password string
}

type UserInfoFromReq map[string]*User

type User struct {
	ID        string
	Email     string
	CreatedAt string
	UpdatedAt string
	Username  string
	Bio       string
	Image     string
	Token     string
	Password  string
	Following bool
}

type UserResp struct {
	Email     string
	CreatedAT FakeTime
	UpdatedAt FakeTime
	UserName  string
}

type UserHandler struct {
	DB map[LoginData]User
}

func CreateUserHandler() *UserHandler {
	return &UserHandler{DB: make(map[LoginData]User)}
}

func CreateUserInfo(map[string]int) {}

func (u *UserHandler) Registartion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := make(UserInfoFromReq)
		BUserData, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}

		err = json.Unmarshal(BUserData, &user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}

		userinfo := user["user"]
		u.DB[LoginData{Login: userinfo.Username, Password: userinfo.Password}] = *userinfo

		userinfo.CreatedAt = "2012-04-23T18:25:43.511Z"
		userinfo.UpdatedAt = "2012-04-23T18:25:43.511Z"

		JSuinfo, _ := json.Marshal(map[string]*User{"User": userinfo})
		fmt.Println(string(JSuinfo))

		w.WriteHeader(201)
		w.Write(JSuinfo)
	}
}

// func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
// 	u.DB[l]
// }

func GetApp() http.Handler {
	mux := http.NewServeMux()
	u := CreateUserHandler()
	mux.HandleFunc("/api/users", u.Registartion)

	return mux
}
