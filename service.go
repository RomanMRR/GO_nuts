package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	_ "github.com/go-playground/validator/v10"

	_ "github.com/lib/pq"

	_ "github.com/Jeffail/gabs/v2"
	stan "github.com/nats-io/stan.go"
)

type Order struct {
	OrderUID    string `json:"order_uid" validate:"required"`
	TrackNumber string `json:"track_number" validate:"required"`
	Entry       string `json:"entry" validate:"required"`
	Delivery    struct {
		Name    string `json:"name"`
		Phone   string `json:"phone"`
		Zip     string `json:"zip"`
		City    string `json:"city"`
		Address string `json:"address"`
		Region  string `json:"region"`
		Email   string `json:"email"`
	} `json:"delivery"`
	Payment struct {
		Transaction  string `json:"transaction"`
		RequestID    string `json:"request_id"`
		Currency     string `json:"currency"`
		Provider     string `json:"provider"`
		Amount       int    `json:"amount"`
		PaymentDt    int    `json:"payment_dt"`
		Bank         string `json:"bank"`
		DeliveryCost int    `json:"delivery_cost"`
		GoodsTotal   int    `json:"goods_total"`
		CustomFee    int    `json:"custom_fee"`
	} `json:"payment"`
	Items []struct {
		ChrtID      int    `json:"chrt_id"`
		TrackNumber string `json:"track_number"`
		Price       int    `json:"price"`
		Rid         string `json:"rid"`
		Name        string `json:"name"`
		Sale        int    `json:"sale"`
		Size        string `json:"size"`
		TotalPrice  int    `json:"total_price"`
		NmID        int    `json:"nm_id"`
		Brand       string `json:"brand"`
		Status      int    `json:"status"`
	} `json:"items"`
	Locale            string    `json:"locale"`
	InternalSignature string    `json:"internal_signature"`
	CustomerID        string    `json:"customer_id"`
	DeliveryService   string    `json:"delivery_service"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

var cache = map[string]Order{}

func CheckError(err error, error_description string) {
	if err != nil {
		fmt.Println(error_description)
		panic(err)
	}
}

func home_page(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/home_page.html", "templates/header.html", "templates/footer.html")
	CheckError(err, "Impossible parse html files")
	tmpl.ExecuteTemplate(w, "index", nil)
}

func show_order(w http.ResponseWriter, r *http.Request) {
	uid := r.FormValue("uid")
	tmpl, err := template.ParseFiles("templates/home_page.html", "templates/header.html", "templates/footer.html", "templates/order.html")
	CheckError(err, "Impossible parse html files")
	v, _ := cache[uid]
	tmpl.ExecuteTemplate(w, "order", v)

}

func handle_requests() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", home_page)
	http.HandleFunc("/show", show_order)
	http.ListenAndServe(":8080", nil)
}

func validate_json(order *Order) bool {
	validate := validator.New()
	err := validate.Struct(order)
	if err != nil {
		// log out this error
		// log.Error(err)
		// return a bad request and a helpful error message
		// if you wished, you could concat the validation error into this
		// message to help point your consumer in the right direction.
		fmt.Println("Invalid JSON")
		return false
	}
	fmt.Println("Validation succesfull!")
	return true
}

func main() {

	sc, _ := stan.Connect("prod", "sub-1", stan.NatsURL(stan.DefaultNatsURL),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Printf("Connection lost, reason: %v", reason)
		}))
	defer sc.Close()
	connStr := "postgresql://roman:1735@127.0.0.1:5432/orders?sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	CheckError(err, "Unable to connect to database")

	recovered_data, err := db.Query("SELECT * FROM order_table")

	CheckError(err, "Unable to recover data from database")

	defer recovered_data.Close()

	for recovered_data.Next() {
		var id string
		var description []byte
		var order Order
		err = recovered_data.Scan(&id, &description)
		CheckError(err, "Unable to get data from database")
		json.Unmarshal(description, &order)
		cache[id] = order
		fmt.Printf("Added %s in cache\n", id)
	}

	var temp Order

	sc.Subscribe("Order", func(m *stan.Msg) {
		temp = Order{}
		err = json.Unmarshal(m.Data, &temp)
		if err != nil {
			log.Printf("Error whilу getting JSON: %s", err.Error())
			return
		}
		if validate_json(&temp) {
			if _, ok := cache[temp.OrderUID]; ok {
				fmt.Printf("%s уже существует\n", cache[temp.OrderUID].OrderUID)
				return
			}
			fmt.Println(temp)
			cache[temp.OrderUID] = temp

			_, err = db.Exec("INSERT INTO order_table (order_uid, description) VALUES ($1, $2)", temp.OrderUID, m.Data)
			if err != nil {
				log.Println(err, "Error while insert JSON into Postgres")
			}

			fmt.Printf("Данные по %s сохранены\n", temp.OrderUID)
		} else {
			fmt.Println("Unable to save data")
		}

	})
	handle_requests()

}
