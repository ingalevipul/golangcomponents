package main

import (
	"fmt"
	"tutproj/internal/model"
	"tutproj/internal/repository/postgres"
)

func main() {
	db, err := postgres.CreateConnection()

	if err != nil {
		fmt.Println("error creating connection")
	}

	//postgres.CreateUser(db)
	//postgres.CreateUserBatch(db)
	postgres.GetUsers(db)
	//postgres.Pagantion(db, 2)
	// var user model.User

	// user = model.User{Name: "viul", Age: 20}
	// postgres.GetRowByfilter(db, user)
	// var searchuser model.User
	// var updateduser model.User
	// searchuser = model.User{Name: "vipul"}
	// updateduser = model.User{Name: "vipul ingale", Age: 20}
	// postgres.UpdateUser(db, searchuser, updateduser)
	//postgres.GetUsers(db)
	// postgres.DeleteUser(db, model.User{Name: "manish"}, true)

	// postgres.GetUsers(db)
	postgres.AddPost(db, model.Post{Title: "hello", PostID: 1})

}
