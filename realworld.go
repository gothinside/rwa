package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func JsonResponse(data interface{}, w http.ResponseWriter, status int) {
	JSONData, _ := json.Marshal(data)
	w.WriteHeader(status)
	w.Write(JSONData)
}

type FakeTime struct {
	Valid bool
}

type ISessionManager interface {
	Create(http.ResponseWriter, *User) error
	Check(*http.Request) (*User, error)
	Delete(string)
}

type Session struct {
	User      *User
	SessionID int
}

type SessionManager struct {
	Sessions map[string]Session
	ID       int
}

func (sm *SessionManager) Create(w http.ResponseWriter, u *User) error {
	//id, _ := strconv.Atoi(u.ID)
	sm.Sessions["Token "+strconv.Itoa(sm.ID+1)] = Session{u, sm.ID + 1}
	u.Token = strconv.Itoa(sm.ID + 1)

	cookie := &http.Cookie{
		Name:    "session_id",
		Value:   strconv.Itoa(sm.ID + 1),
		Expires: time.Now().Add(90 * 24 * time.Hour),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	sm.ID++
	return nil
}

func (sm *SessionManager) Check(r *http.Request) (*User, error) {
	if val, ok := sm.Sessions[r.Header.Get("Authorization")]; !ok {
		return nil, fmt.Errorf("User not found")
	} else {
		return val.User, nil
	}
}

type LoginData struct {
	Email    string
	Password string
}

func UnmarshalJsonData(data map[string]interface{}, Body io.ReadCloser) (map[string]interface{}, error) {
	BinaryData, err := io.ReadAll(Body)
	Body.Close()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(BinaryData, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

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

type Author struct {
	Username string
	Bio      string
}

type Article struct {
	Author         Author
	Body           string
	CreatedAt      string
	Description    string
	Favorited      bool
	FavoritesCount int
	Slug           string
	TagList        []string
	Title          string
	UpdatedAt      string
}

func (u *User) ChangeUserData(m map[string]string) *User {
	for key, value := range m {
		switch key {
		case "bio":
			u.Bio = value
		case "email":
			u.Email = value
		case "password":
			u.Password = value
		case "username":
			u.Username = value
		}
	}
	return u
}

type UserHandler struct {
	DB map[LoginData]*User
	SM ISessionManager
	AM *ArticleHandler
}

func (u *UserHandler) FindUserById(ID int) *User {
	for _, value := range u.DB {
		if value.ID == strconv.Itoa(ID) {
			return value
		}
	}
	return nil
}

func (u *UserHandler) AddNewUser(UserInfo *User) {
	fmt.Println(LoginData{Email: UserInfo.Email, Password: UserInfo.Password})
	u.DB[LoginData{Email: UserInfo.Email, Password: UserInfo.Password}] = UserInfo
}

func (u *UserHandler) CheckUser(login LoginData) (*User, error) {
	if _, ok := u.DB[login]; ok {
		return u.DB[login], nil
	} else {
		return nil, fmt.Errorf("incorrect data")
	}
}

func CreateUserHandler() *UserHandler {
	return &UserHandler{
		DB: make(map[LoginData]*User),
		SM: &SessionManager{Sessions: make(map[string]Session)},
		AM: &ArticleHandler{Articles: make(map[string][]*Article)}}
}

func (u *UserHandler) Registartion(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		user := make(map[string]*User)
		BUserData, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		err = json.Unmarshal(BUserData, &user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		UserInfo := user["user"]
		u.AddNewUser(UserInfo)
		UserInfo.CreatedAt = time.Now().Format(time.RFC3339)
		UserInfo.UpdatedAt = time.Now().Format(time.RFC3339)

		err = u.SM.Create(w, UserInfo)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		JsonResponse(map[string]*User{"User": UserInfo}, w, 201)
	}
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	login := make(map[string]LoginData)

	BUserData, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	err = json.Unmarshal(BUserData, &login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	user, err := u.CheckUser(login["user"])
	fmt.Println(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = u.SM.Create(w, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	JsonResponse(map[string]*User{"User": user}, w, 200)
}

func (u *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		user, err := u.SM.Check(r)
		fmt.Println(user)
		if err != nil {
			http.Error(w, err.Error(), 401)
		} else {
			JSUser, _ := json.Marshal(map[string]*User{"User": user})
			fmt.Println(string(JSUser))
			w.Write(JSUser)
		}
	}
	if r.Method == "PUT" {
		user, err := u.SM.Check(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			NewUserInfo := make(map[string]map[string]string)
			BUserData, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			err = json.Unmarshal(BUserData, &NewUserInfo)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
			}

			NewUser := NewUserInfo["user"]
			user := user.ChangeUserData(NewUser)
			JsonResponse(map[string]*User{"User": user}, w, 200)
		}
	}
}

type ArticleHandler struct {
	Articles map[string][]*Article
	AID      int
}

func (AH *ArticleHandler) AddNewArticle(art *Article) {
	fmt.Println(AH.Articles)
	if _, ok := AH.Articles[art.Author.Username]; !ok {
		AH.Articles[art.Author.Username] = []*Article{art}
	} else {
		AH.Articles[art.Author.Username] = append(AH.Articles[art.Author.Username], art)
	}

}

func (u *UserHandler) NewPost(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ArticleData := make(map[string]*Article)
		BArticleData, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		err = json.Unmarshal(BArticleData, &ArticleData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		NewArticle := ArticleData["article"]
		AuthorInfo, err := u.SM.Check(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if AuthorInfo.Bio != "" {
			NewArticle.Author.Bio = AuthorInfo.Bio
		}
		NewArticle.Author.Username = AuthorInfo.Username
		NewArticle.CreatedAt = time.Now().Format(time.RFC3339)
		NewArticle.UpdatedAt = time.Now().Format(time.RFC3339)
		NewArticle.Slug = strconv.Itoa(u.AM.AID)
		u.AM.AID++
		u.AM.AddNewArticle(NewArticle)

		JsonResponse(map[string]*Article{"Article": NewArticle}, w, 201)
	}
	if r.Method == "GET" {
		q := r.URL.Query()
		if AuthorName := q.Get("author"); AuthorName != "" {
			a := u.AM.Articles[AuthorName]
			JSArt, _ := json.Marshal(AllArtResp{Articles: a, ArticlesCount: len(u.AM.Articles[AuthorName])})
			w.WriteHeader(200)
			w.Write(JSArt)
		} else if tag := q.Get("tag"); tag != "" {
			a := []*Article{}
			count := 0
			for _, arts := range u.AM.Articles {
				for _, art := range arts {
					for _, arttag := range art.TagList {
						if arttag == tag {
							a = append(a, art)
							count++
							break
						}
					}
				}
			}
			JsonResponse(AllArtResp{Articles: a, ArticlesCount: count}, w, 200)
		} else {
			a := []*Article{}
			count := 0
			for _, value := range u.AM.Articles {
				a = append(a, value...)
				count++
			}
			JsonResponse(AllArtResp{Articles: a, ArticlesCount: count}, w, 200)
		}
	}
}

type AllArtResp struct {
	Articles      []*Article `json:"articles"`
	ArticlesCount int        `json:"articlesCount"`
}

func (AH *ArticleHandler) GetArticles(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		q := r.URL.Query()
		if q.Get("Author") != "" {
		} else if q.Get("Tag") != "" {

		} else {
			a := []*Article{}
			count := 0
			for _, value := range AH.Articles {
				a = append(a, value...)
				count++
			}
			JsonResponse(AllArtResp{Articles: a, ArticlesCount: count}, w, 200)
		}
	}
}

func (SM *SessionManager) Delete(Token string) {
	delete(SM.Sessions, Token)
}

func (u *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	user, err := u.SM.Check(r)
	if err != nil {
		http.Error(w, err.Error(), 401)
	} else {
		u.SM.Delete("Token " + user.Token)
	}
}

func GetApp() http.Handler {
	mux := http.NewServeMux()
	u := CreateUserHandler()
	mux.HandleFunc("/api/users", u.Registartion)
	mux.HandleFunc("/api/users/login", u.Login)
	mux.HandleFunc("/api/user", u.Profile)
	mux.HandleFunc("/api/articles", u.NewPost)
	mux.HandleFunc("/api/user/logout", u.Logout)
	return mux
}
