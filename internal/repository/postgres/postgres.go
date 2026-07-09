package postgres

import (
	"fmt"
	"log"
	"os"
	"tutproj/internal/model"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func CreateLogin(db *gorm.DB, user model.UnamePass) error {
	err := db.Create(&user).Error
	if err != nil {
		fmt.Println("error creating username with name: ", user.Username)
		return err
	}
	return nil

}
func AddPost(db *gorm.DB, newpost model.Post) {
	var user model.User
	db.First(&user, 1)

	err := db.Model(&user).Association("Posts").Append(&newpost)

	if err != nil {
		fmt.Println("Error adding post:", err)
	} else {
		fmt.Println("Successfully added post for user:", user.Name)
	}
}
func UpdateUser(db *gorm.DB, searchuser model.User, updateduser model.User) {
	var userToUpdate model.User
	result := db.Where(&searchuser).First(&userToUpdate)

	if result.Error != nil {
		fmt.Println("Error finding user:", result.Error)
		return
	}

	db.Model(&userToUpdate).Updates(updateduser)
}

func DeleteUser(db *gorm.DB, user model.User, perm bool) {
	db.Where(&user).Delete(&user)
	if perm {
		db.Unscoped().Where(&user).Delete(&user)
	}

}
func CreateUserBatch(db *gorm.DB) {
	userlist := []model.User{
		{Name: "vipul", Email: "vipul.ingale147@gmail.com", Password: "vipul1", Status: "inactive", Age: 20},
		{Name: "manish", Email: "manishaladwani@gmail.com", Password: "manishl", Age: 19},
		{Name: "rahul", Email: "rahul.sharma@gmail.com", Password: "rahul123", Status: "active", Age: 25},
		{Name: "priya", Email: "priya.patel@gmail.com", Password: "priya456", Status: "active", Age: 22},
		{Name: "amit", Email: "amit.kumar@gmail.com", Password: "amit789", Status: "inactive", Age: 30},
		{Name: "sneha", Email: "sneha.joshi@gmail.com", Password: "sneha321", Status: "active", Age: 27},
		{Name: "rohit", Email: "rohit.verma@gmail.com", Password: "rohit654", Status: "active", Age: 23},
		{Name: "anjali", Email: "anjali.singh@gmail.com", Password: "anjali987", Status: "inactive", Age: 21},
		{Name: "vikram", Email: "vikram.nair@gmail.com", Password: "vikram111", Status: "active", Age: 35},
		{Name: "pooja", Email: "pooja.desai@gmail.com", Password: "pooja222", Status: "active", Age: 28},
		{Name: "arjun", Email: "arjun.mehta@gmail.com", Password: "arjun333", Status: "active", Age: 26},
		{Name: "kavya", Email: "kavya.reddy@gmail.com", Password: "kavya444", Status: "inactive", Age: 24},
		{Name: "suresh", Email: "suresh.iyer@gmail.com", Password: "suresh555", Status: "active", Age: 40},
		{Name: "divya", Email: "divya.pillai@gmail.com", Password: "divya666", Status: "active", Age: 29},
		{Name: "nikhil", Email: "nikhil.gupta@gmail.com", Password: "nikhil777", Status: "active", Age: 32},
		{Name: "meera", Email: "meera.khanna@gmail.com", Password: "meera888", Status: "inactive", Age: 19},
		{Name: "karan", Email: "karan.malhotra@gmail.com", Password: "karan999", Status: "active", Age: 33},
		{Name: "tanya", Email: "tanya.bose@gmail.com", Password: "tanya000", Status: "active", Age: 31},
		{Name: "aakash", Email: "aakash.rao@gmail.com", Password: "aakash123", Status: "inactive", Age: 22},
		{Name: "ritu", Email: "ritu.saxena@gmail.com", Password: "ritu456", Status: "active", Age: 27},
	}
	result := db.CreateInBatches(userlist, 10)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	fmt.Println("inserted data")

}
func GetUsers(db *gorm.DB) {

	var user []model.User
	db.Find(&user)

	for _, u := range user {
		fmt.Println(u.Name)
	}

}
func Pagantion(db *gorm.DB, pgnum int) {
	var user []model.User
	limit := 10
	offest := (pgnum - 1) * limit
	db.Offset(offest).Limit(limit).Find(&user)

	for _, u := range user {
		fmt.Println(u.Name, u.Email, u.Status)
	}

}
func GetRowByfilter(db *gorm.DB, user model.User) {
	var res []model.User
	result := db.Where(user).Find(&res)
	if result.Error != nil {
		fmt.Println(result.Error)
	}
	for _, u := range res {
		fmt.Println(u)
	}

}

func CreateUser(db *gorm.DB) {

	user := model.User{Name: "vipul", Email: "vipul.ingale147@gmail.com", Password: "hello", Age: 19}

	result := db.Create(&user)

	fmt.Println(result.RowsAffected)

}
func CreateConnection() (*gorm.DB, error) {
	godotenv.Load("/Users/vipulingale/Desktop/tutproj/.env")
	url := os.Getenv("POSTGRES_URL")
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\n", log.LstdFlags),
			logger.Config{},
		),
	})

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	fmt.Println("connection eastablished")

	err = db.AutoMigrate(&model.User{}, &model.Post{}, &model.Language{}, &model.UnamePass{})
	if err != nil {
		fmt.Println("Error auto-migrating schema:", err)
		return nil, err
	}

	fmt.Println("migrated all tables")
	return db, nil
}
