package database

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/Strike-official/reddeggsBot/schema"
	_ "github.com/go-sql-driver/mysql"
)

func ConnectToRDS() *sql.DB {
	db, err := sql.Open("mysql", "")

	if err != nil {
		log.Println(err.Error())
	}
	return db
}

func GetUserRDS(db *sql.DB, request schema.Strike_Meta_Request_Structure) string {
	bybrsik_id := request.Bybrisk_session_variables.UserId
	results, err := db.Query("SELECT user_id, name FROM User where bybrisk_id='" + bybrsik_id + "'")
	if err != nil {
		log.Println(err.Error())
	}

	type wrapper struct {
		Name       string `json:"name"`
		User_id    string `json:"user_id"`
		Bybrisk_id string `json:"bybrisk_id"`
	}
	var wrapperObj wrapper
	for results.Next() {
		err = results.Scan(&wrapperObj.User_id, &wrapperObj.Name)
		if err != nil {
			log.Println(err.Error())
		}
	}

	if wrapperObj.User_id == "" {
		wrapperObj.User_id = AddUserRDS(db, request)
		log.Println("New User Added! " + wrapperObj.User_id)
		return wrapperObj.User_id
	}
	log.Println("User Found! " + wrapperObj.User_id)
	return wrapperObj.User_id
}

func AddUserRDS(db *sql.DB, request schema.Strike_Meta_Request_Structure) string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		log.Println(err.Error())
	}
	id := fmt.Sprintf("%X", b)

	name := request.Bybrisk_session_variables.Username
	bybrsik_id := request.Bybrisk_session_variables.UserId
	_, err := db.Query("INSERT INTO User VALUES ( '" + id + "', '" + name + "','" + bybrsik_id + "')")

	if err != nil {
		log.Println("[DB ERROR] Error Adding user: ", err.Error())
	}
	return id
}

func AddOrder(db *sql.DB, request schema.Strike_Meta_Request_Structure, id string, item_total int64, item_description []string, delivery_date string) (string, error) {

	phone := request.Bybrisk_session_variables.Phone
	latitude := request.Bybrisk_session_variables.Location.Latitude
	// if latitude, err := strconv.ParseFloat(latitude_string, 64); err != nil {
	// 	log.Println(err)
	// }

	longitude := request.Bybrisk_session_variables.Location.Longitude
	// if longitude, err := strconv.ParseFloat(longitude_string, 64); err != nil {
	// 	log.Println(err)
	// }

	address := request.Bybrisk_session_variables.Address

	var final_item_desc string
	for index, item_description_indi := range item_description {
		if index == len(item_description)-1 {
			final_item_desc = final_item_desc + strings.TrimSpace(item_description_indi)
		} else {
			final_item_desc = final_item_desc + strings.TrimSpace(item_description_indi) + ", "
		}
	}

	quantity := request.User_session_variables.OrderQuantity
	status := "ACTIVE"

	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "E1DB", err
	}
	order_id := fmt.Sprintf("%X", b)

	item_total_double := float64(item_total)

	_, err := db.Query("INSERT INTO Order_detail (user_id,phone,latitude,longitude,address,item_description,quantity,status,order_id,item_total,delivery_date) VALUES ('" + id + "','" + phone + "'," + fmt.Sprintf("%v", latitude) + "," + fmt.Sprintf("%v", longitude) + ",'" + address + "','" + final_item_desc + "','" + quantity + "','" + status + "','" + order_id + "'," + fmt.Sprintf("%v", item_total_double) + ",'" + delivery_date + "')")

	if err != nil {
		return "E2DB", err
	}
	return order_id, nil
}

func GetOrders(db *sql.DB, request schema.Strike_Meta_Request_Structure) []schema.Wrapper {
	bybrisk_id_val := request.Bybrisk_session_variables.UserId
	results, err := db.Query("select item_description,item_total,quantity,order_time,delivery_date from Order_detail where user_id=(select user_id from User where bybrisk_id='" + bybrisk_id_val + "') order by order_time desc")
	if err != nil {
		log.Println(err.Error())
	}

	var wrapperObjArray []schema.Wrapper

	for results.Next() {
		var wrapperObj schema.Wrapper
		err = results.Scan(&wrapperObj.Item_description, &wrapperObj.Item_total, &wrapperObj.Quantity, &wrapperObj.Order_time, &wrapperObj.Delivery_date)
		if err != nil {
			log.Println(err.Error())
		}
		wrapperObjArray = append(wrapperObjArray, wrapperObj)
	}
	return wrapperObjArray
}
