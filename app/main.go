package main

import (
	"github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
	"gopkg.in/mgo.v2"
	"log"
	"net/http"
)

const (
	//애플리케이션에서 사용할 세션의 키 정보
	sessionKey    = "go_chatting_session"
	sessionSecret = "go_chatting_session_secret"
	socketBufferSize=1024
)

var (
	renderer     *render.Render
	mongoSession *mgo.Session

	upgrader=&websocket.Upgrader{
		ReadBufferSize: socketBufferSize,
		WriteBufferSize: socketBufferSize,
	}
)

func init() {
	//렌더러 생성
	renderer = render.New()

	s,err:=mgo.Dial("mongodb://localhost")
	if err!=nil{
		panic(err)
	}

	mongoSession=s
}

func main() {

	//TODO: room 삭제 시 DB에 저장된 메시지도 같이 삭제.
	//TODO: room 생성 시 room name 중복 불가 처리.
	//TODO: logout 추가.
	//TODO: auth.go  nextPageKey 사용 수정 필요.

	//라우터 생성
	router := httprouter.New()

	router.GET("/", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		//렌더러를 사용하여 템플릿 렌더링
		renderer.HTML(w, http.StatusOK, "index", map[string]interface{}{"host": r.Host})
	})

	router.GET("/login", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		//로그인 페이지 렌더링
		renderer.HTML(w, http.StatusOK, "login", nil)
	})

	router.GET("/logout", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		//세션에서 사용자 정보 제거 후 로그인 페이지로 이동
		sessions.GetSession(r).Delete(currentUserKey)
		http.Redirect(w, r, "/login", http.StatusFound)
	})

	router.GET("/auth/:action/:provider", loginHandler)

	router.POST("/rooms/create",createRoom)
	router.GET("/rooms",retrieveRooms)

	router.GET("/rooms/:id/messages",retrieveMessages)

	router.GET("/ws/:room_id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		socket, err := upgrader.Upgrade(w,r,nil)
		if err!=nil{
			log.Fatal("ServeHTTP:",err)
			return
		}
		newClient(socket,ps.ByName("room_id"),GetCurrentUser(r))
	})

	router.POST("/rooms/delete",deleteRoom)

	//negroni 미들웨어 생성
	n := negroni.Classic()
	store := cookiestore.New([]byte(sessionSecret))
	n.Use(sessions.Sessions(sessionKey, store))
	n.Use(LoginRequired("/login", "/auth"))

	//negroni에 router를 핸들러로 등록
	n.UseHandler(router)

	//웹 서버 실행
	n.Run(":3000")

}
