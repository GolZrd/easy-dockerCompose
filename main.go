package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Product struct {
	Id      int
	Model   string
	Company string
	Price   int
}

var database *gorm.DB

// Функция удаления записи из БД
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	result := database.Delete(&Product{}, id)
	if result.Error != nil {
		log.Println(result.Error)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

// возвращаем пользователю страницу для редактирования объекта
func EditPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	prod := Product{}

	// Есть 2 варианта как можно получить данные из БД и вставить в поля структуры.
	// Первый вариант исопльзуя метод Table (для того чтобы выбрать бд из которой брать данные), Select (для выбора полей), Where (для выбора записей по id) и Scan (для вставки полученных данных в структуру)
	err := database.Table("products").Select("id", "model", "company", "price").Where("id = ?", id).Scan(&prod)
	// Второй вариант это использование Raw, где мы вставляем sql запрос и получаем данные с помощью Scan
	//err := database.Raw("SELECT * FROM products WHERE id = ?", id).Scan(&prod)
	if err == nil {
		log.Println(err)
		http.Error(w, http.StatusText(404), http.StatusNotFound)
	} else {
		tmpl, _ := template.ParseFiles("templates/edit.html")
		tmpl.Execute(w, prod)
	}
}

// получаем измененные данные и сохраняем их в БД
func EditHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
	}
	id, err := strconv.Atoi(r.FormValue("id"))
	fmt.Println(id)
	if err != nil {
		log.Fatal(err)
	}
	model := r.FormValue("model")
	company := r.FormValue("company")
	price, err := strconv.Atoi(r.FormValue("price"))
	if err != nil {
		log.Fatal(err)
	}
	newProduct := Product{Id: id, Model: model, Company: company, Price: price}
	result := database.Save(&newProduct)
	if result.Error != nil {
		fmt.Errorf("Failed to create product: %v", result.Error)
	}
	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println(err)
		}

		model := r.FormValue("model")
		company := r.FormValue("company")
		price, err := strconv.Atoi(r.FormValue("price"))
		if err != nil {
			log.Fatal(err)
		}
		newProduct := Product{Model: model, Company: company, Price: price}
		result := database.Create(&newProduct)
		if result.Error != nil {
			log.Println(result.Error)
			http.Error(w, http.StatusText(404), http.StatusNotFound)
			return
		}
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
	} else {
		http.ServeFile(w, r, "templates/create.html")
	}

}

func IndexHandler(w http.ResponseWriter, r *http.Request) {

	var products []Product
	database.Find(&products)

	tmpl, _ := template.ParseFiles("templates/index.html")
	tmpl.Execute(w, products)
}

func main() {
	connStr := "user=postgres password=Egolgor23 dbname=productdb sslmode=disable"
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Product{})
	if err != nil {
		log.Fatal(err)
	}

	database = db

	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/create", CreateHandler)
	router.HandleFunc("/edit/{id:[0-9]+}", EditPage).Methods("GET")
	router.HandleFunc("/edit/{id:[0-9]+}", EditHandler).Methods("POST")
	router.HandleFunc("/delete/{id:[0-9]+}", DeleteHandler)

	http.Handle("/", router)

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(":8181", router))
}
