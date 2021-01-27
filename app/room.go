package main

import (
	"github.com/julienschmidt/httprouter"
	"github.com/mholt/binding"
	"gopkg.in/mgo.v2/bson"
	"net/http"
)

type Room struct{
	ID bson.ObjectId `bson:"_id" json:"id"`
	Name string `bson:"name" json:"name"`
}

func (r* Room) FieldMap(req *http.Request) binding.FieldMap{
	return binding.FieldMap{&r.Name:"name"}
}

func createRoom(w http.ResponseWriter, req *http.Request,ps httprouter.Params){
	//binding 패키지로 room 생성 요청 정보를 Room 타입 값으로 변환
	r:=new(Room)
	errs := binding.Bind(req,r)
	if errs != nil{
		return
	}

	//몽고DB 세션 생성
	session := mongoSession.Copy()
	//몽고DB 세션을 닫는 코드를 defer로 등록
	defer session.Close()

	//몽고DB 아이디 생성
	r.ID=bson.NewObjectId()
	//room 정보 저장을 위한 몽고DB 컬렉션 객체 생성
	c:=session.DB("test").C("rooms")

	//rooms 컬렉션에 room 정보 정장
	if err:=c.Insert(r); err!=nil{
		//오류 발생 시 500 에러 반환
		renderer.JSON(w,http.StatusInternalServerError,err)
		return
	}

	//처리 결과 반환
	renderer.JSON(w,http.StatusCreated,r)
}

func retrieveRooms(w http.ResponseWriter,req *http.Request,ps httprouter.Params){
	//몽고DB 세션 생성
	session:=mongoSession.Copy()
	//몽고DB 세션을 닫는 코드를 defer로 등록록
	defer session.Close()

	var rooms []Room
	//모든 room 정보 조회
	err := session.DB("test").C("rooms").Find(nil).All(&rooms)
	if err!=nil {
		//오류 발생 시 500 에러 반환
		renderer.JSON(w,http.StatusInternalServerError,err)
		return
	}

	//room 조회 결과 반환
	renderer.JSON(w,http.StatusOK,rooms)
}

//func retrieveRoom(w http.ResponseWriter,req *http.Request,ps httprouter.Params){
//	session := mongoSession.Copy()
//	defer session.Close()
//
//	var room Room
//	err:=session.DB("test").C("rooms").FindId(bson.ObjectIdHex(ps.ByName("id"))).One(&room)
//	if err!=nil{
//		renderer.JSON(w,http.StatusInternalServerError,err)
//		return
//	}
//
//	renderer.JSON(w,http.StatusOK,room)
//}

func deleteRoom(w http.ResponseWriter, req *http.Request,ps httprouter.Params){
	session:=mongoSession.Copy()
	defer session.Close()

	r:=new(Room)
	//바인딩
	errs := binding.Bind(req,r)
	if errs != nil{
		return
	}

	//room 이름으로 데이터베이스에서 찾고 r에 해당 room 정보 저장.
	err:=session.DB("test").C("rooms").Find(bson.M{"name":r.Name}).One(&r)

	if err!=nil{
		renderer.JSON(w,http.StatusInternalServerError,err)
		return
	}

	//해당 room에 연결된 messages들 전부 삭제.
	_,err = session.DB("test").C("messages").RemoveAll(bson.M{"room_id":r.ID})

	if err!=nil{
		renderer.JSON(w,http.StatusInternalServerError,err)
		return
	}

	//마지막으로 room 삭제.
	err = session.DB("test").C("rooms").RemoveId(r.ID)
	//err := session.DB("test").C("rooms").Remove(bson.M{"name":r.Name})

	if err!=nil{
		renderer.JSON(w,http.StatusInternalServerError,err)
		return
	}

	renderer.JSON(w,http.StatusNoContent,nil)
}