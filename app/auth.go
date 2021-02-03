package main

import (
	"github.com/codegangsta/negroni"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"log"
	"net/http"
	"strings"
)

const (
	nextPageKey     = "next_page" //세션에 저장되는 next page의 키
	authSecurityKey = "auth_security_key"
)

func init() {
	//gomniauth 정보 세팅
	gomniauth.SetSecurityKey(authSecurityKey)
	gomniauth.WithProviders(
		// 사용 시 clientId와 clientSecret 입력 필요.
		google.New("clientID", "ClientIDSecret", "http://127.0.0.1:3000/auth/callback/google"),
	)
}

func loginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	action := ps.ByName("action")
	provider := ps.ByName("provider")
	//세션에 저장한 nextPageKey를 받아올수가 없으므로 당장은 필요하지 않음. 추후 수정 필요.
	//s := sessions.GetSession(r)

	switch action {
	case "login":
		//gomniauth.Provider의 login 페이지로 이동
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln(err)
		}
		loginUrl, err := p.GetBeginAuthURL(nil, nil)
		if err != nil {
			log.Fatalln(err)
		}
		http.Redirect(w, r, loginUrl, http.StatusFound)
	case "callback":
		//gomniauth 콜백 처리
		p, err := gomniauth.Provider(provider)
		if err != nil {
			log.Fatalln(err)
		}
		creds, err := p.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))
		if err != nil {
			log.Fatalln(err)
		}

		//콜백 결과로부터 사용자 정보 확인
		user, err := p.GetUser(creds)
		if err != nil {
			log.Fatalln(err)
		}

		if err != nil {
			log.Fatalln(err)
		}

		u := &User{
			Uid:       user.Data().Get("id").MustStr(),
			Name:      user.Name(),
			Email:     user.Email(),
			AvatarUrl: user.AvatarURL(),
		}

		SetCurrentUser(r, u)//사용자 정보를 세션에 저장
		//LoginRequired에서 nextPageKey로 set한 정보를 get으로 안 받아짐. 추후 수정 필요.
		//http.Redirect(w, r, s.Get(nextPageKey).(string), http.StatusFound)
		http.Redirect(w,r,"/",http.StatusFound)
	default:
		http.Error(w, "Auth action `"+action+"' is not supported", http.StatusNotFound)
	}
}

func LoginRequired(ignore ...string) negroni.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		//ignore url이면 다음 핸들러 실행
		for _, s := range ignore {
			if strings.HasPrefix(r.URL.Path, s) {
				next(w, r)
				return
			}
		}
		//Currentuser 정보를 가져옴
		u := GetCurrentUser(r)

		//CurrentUser 정보가 유효하면 만료 시간을 갱신하고 다음 핸들러 실행
		if u != nil && u.Valid() {
			SetCurrentUser(r, u)
			next(w, r)
			return
		}

		//CurrentUser 정보가 유효하지 않으면 CurrentUser를 nil로 세팅
		SetCurrentUser(r, nil)

		//로그인 후 이동할 때 url을 세션에 저장(r)
		//sessions.GetSession(r).Set(nextPageKey, r.URL.RequestURI())
		//저장한 url을 loginHandler에서 확인할 수 없음. 추후 확인 필요.

		//로그인 페이지로 리다이렉트
		http.Redirect(w, r, "/login", http.StatusFound)

	}
}
