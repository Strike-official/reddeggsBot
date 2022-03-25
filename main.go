package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/strike-official/go-sdk/strike"

	"strconv"

	"github.com/Strike-official/reddeggsBot/database"
	"github.com/Strike-official/reddeggsBot/schema"
	"github.com/Strike-official/reddeggsBot/telegram"
	"github.com/shashank404error/draco/mongo"
)

var (
	base_url         = ""
	TelegramAPIToken = ""
)

func determineListenAddress() (string, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6005"
	}
	return ":" + port, nil
}

func main() {
	addr, err := determineListenAddress()
	if err != nil {
		log.Fatal(err)
	}
	mongo.MongoInit()
	r := mux.NewRouter()
	//r.HandleFunc("/",index).Methods("POST")
	r.HandleFunc("/reddeggs/order-view", order_view).Methods("POST")
	r.HandleFunc("/reddeggs/order-controller", order_controller).Methods("POST")
	r.HandleFunc("/reddeggs/order-history", order_history).Methods("POST")
	r.HandleFunc("/reddeggs/redirect", redirect).Methods("POST")
	http.Handle("/reddeggs/", r)

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

type productDesc struct {
	PicURL string
	Price  int64
}

var product_name_pic map[string]productDesc = map[string]productDesc{
	"    30 Eggs Tray": productDesc{
		PicURL: "https://raw.githubusercontent.com/shashank404error/reddeggs/master/images/600f2139-b74e-4ba9-bdaf-4944d5b9808c-removebg-preview%20(1).png",
		Price:  190,
	},
	"    12 Pack Eggs": productDesc{
		PicURL: "https://raw.githubusercontent.com/shashank404error/reddeggs/master/images/0c6e5ed4-6e8e-40f2-8f14-ce8c58e74c66-removebg-preview.png",
		Price:  90,
	},
}

//#####################################################HANDLERS###########################################################
func order_view(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, GetOrderData(DecodeRequestToStruct(r.Body)))
}

func order_controller(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, PlaceOrder(DecodeRequestToStruct(r.Body)))
}

func order_history(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, GetHistoryById(DecodeRequestToStruct(r.Body)))
}

func redirect(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, RoutingToAction(DecodeRequestToStruct(r.Body)))
}

//######################################################CORE#####################################################3

func GetOrderData(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	name := request.Bybrisk_session_variables.Username
	strike_object := strike.Create("order-controller", base_url+"/order-controller")

	question_object1 := strike_object.Question("order_type").
		QuestionText().SetTextToQuestion("Great "+name+"! let's select a type", "Text Description, getting used for testing purpose.")

	answer_object1 := question_object1.Answer(true).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)

	answer_object1.AnswerCard().SetHeaderToAnswer(2, strike.HALF_WIDTH).
		AddGraphicRowToAnswer(strike.PICTURE_ROW, []string{"https://raw.githubusercontent.com/shashank404error/reddeggs/master/images/600f2139-b74e-4ba9-bdaf-4944d5b9808c-removebg-preview%20(1).png"}, []string{}).
		AddTextRowToAnswer(strike.H4, "    30 Eggs Tray", "black", true).
		AddTextRowToAnswer(strike.H4, "       ₹ 190", "#009646", true).
		AddTextRowToAnswer(strike.H4, "       ̶₹̶ 2̶1̶0̶", "#ed0212", true).
		AnswerCard().SetHeaderToAnswer(2, strike.HALF_WIDTH).
		AddGraphicRowToAnswer(strike.PICTURE_ROW, []string{"https://raw.githubusercontent.com/shashank404error/reddeggs/master/images/0c6e5ed4-6e8e-40f2-8f14-ce8c58e74c66-removebg-preview.png"}, []string{}).
		AddTextRowToAnswer(strike.H4, "    12 Pack Eggs", "black", true).
		AddTextRowToAnswer(strike.H4, "       ₹ 90", "#009646", true)

	question_object2 := strike_object.Question("order_quantity").
		QuestionText().SetTextToQuestion("How many trays do you want?", "Text Description, getting used for testing purpose.")

	question_object2.Answer(false).NumberInput("Get quantity")

	question_object3 := strike_object.Question("order_date").
		QuestionText().SetTextToQuestion("When should we deliver?", "Text Description, getting used for testing purpose.")

	question_object3.Answer(false).DateInput("Get delivery date")

	return strike_object
}

func PlaceOrder(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	var totalPrice int64
	for _, product := range request.User_session_variables.OrderType {
		totalPrice = totalPrice + product_name_pic[product].Price
	}
	quantity_int64, err := strconv.ParseInt(request.User_session_variables.OrderQuantity, 10, 64)
	if err != nil {
		fmt.Println("error!")
	}

	strike_object := strike.Create("controller", base_url+"/redirect")

	if quantity_int64 > 11 {
		strike_object = strike.Create("order-controller", base_url+"/order-view")
		question_object := strike_object.Question("order_type").
			QuestionText().SetTextToQuestion("Order quantity must be below 10 trays", "Text Description, getting used for testing purpose.")

		answer_object := question_object.Answer(false).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)

		answer_object.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "Place order again!", "black", true)
	} else {

		totalPrice = totalPrice * quantity_int64

		//add order to db
		db := database.ConnectToRDS()
		defer db.Close()
		id := database.GetUserRDS(db, request)

		// Send Telegram notification to listners

		order_ID, db_err := database.AddOrder(db, request, id, totalPrice, request.User_session_variables.OrderType, request.User_session_variables.OrderDate[0])
		if db_err != nil {
			log.Println("[DB ERROR] Error Adding Order: ", db_err)
		}
		telegram.PushToTelegram(TelegramAPIToken, request, request.User_session_variables.OrderType, totalPrice, request.User_session_variables.OrderDate[0], order_ID)

		//construct response
		question_object1 := strike_object.Question("route").
			QuestionCard().SetHeaderToQuestion(1, strike.HALF_WIDTH).AddTextRowToQuestion(strike.H4, "Your Order is confirmed!", "black", true).
			AddTextRowToQuestion(strike.H4, "Item Total ₹ "+strconv.FormatInt(totalPrice, 10), "green", true)

		answer_object1 := question_object1.Answer(false).AnswerCardArray(strike.VERTICAL_ORIENTATION)

		for _, product := range request.User_session_variables.OrderType {
			answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, strike.HALF_WIDTH).
				AddGraphicRowToAnswer(strike.PICTURE_ROW, []string{product_name_pic[product].PicURL}, []string{}).
				AddTextRowToAnswer(strike.H4, product, "black", true).
				AddTextRowToAnswer(strike.H4, strconv.FormatInt(product_name_pic[product].Price, 10)+" * "+request.User_session_variables.OrderQuantity, "#009646", true)

		}

		answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "Expect delivery between 9AM - 7PM "+request.User_session_variables.OrderDate[0], "black", true).
			AddTextRowToAnswer(strike.H4, "Call on 9340212623 to modify time", "black", false)

		answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "Place order again🥚", "#3c6dbd", false)
		answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "View order history", "#3c6dbd", false)
	}
	return strike_object
}

func GetHistoryById(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	db := database.ConnectToRDS()
	defer db.Close()
	orders := database.GetOrders(db, request)

	name := request.Bybrisk_session_variables.Username

	strike_object := strike.Create("order_history", base_url+"/redirect")

	question_object1 := strike_object.Question("route").
		QuestionCard().SetHeaderToQuestion(1, strike.HALF_WIDTH).AddTextRowToQuestion(strike.H4, "Hi "+name+", here's your order history", "#0088e3", true).
		AddTextRowToQuestion(strike.H4, "Select an order to take action", "black", false)

	answer_object1 := question_object1.Answer(false).AnswerCardArray(strike.VERTICAL_ORIENTATION)

	if len(orders) == 0 {
		answer_object1.AnswerCard().SetHeaderToAnswer(10, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H2, "No orders yet!", "black", true)
	} else {
		for _, order := range orders {
			answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, strike.HALF_WIDTH).
				AddTextRowToAnswer(strike.H4, order.Quantity+" × "+order.Item_description, "#474747", true).
				AddTextRowToAnswer(strike.H4, "₹ "+fmt.Sprintf("%v", order.Item_total), "#009646", true).
				AddTextRowToAnswer(strike.H5, "Orderd on: "+order.Order_time.Format("02 Jan 2006"), "black", true)

			now := time.Now().Format("02 Jan 2006")
			if order.Delivery_date == now {
				answer_object1.AddTextRowToAnswer(strike.H5, "Delivery today!", "#de680d", true)
			} else {
				timeStampString := order.Delivery_date
				layOut := "02 Jan 2006"
				timeStamp, err := time.Parse(layOut, timeStampString)
				if err != nil {
					log.Println("Error parsing timestamp: ", err)
				}

				// Delivery in future
				if timeStamp.Unix() > time.Now().Unix() {
					answer_object1.AddTextRowToAnswer(strike.H5, "delivery on: "+order.Delivery_date, "#009646", true)
				} else {
					answer_object1.AddTextRowToAnswer(strike.H5, "delivered on: "+order.Delivery_date, "#ff002b", true)
				}
			}

		}
	}
	answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
		AddTextRowToAnswer(strike.H4, "Place order again🥚", "#3c6dbd", false)
	answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
		AddTextRowToAnswer(strike.H4, "View order history", "#3c6dbd", false)
	return strike_object
}

func RoutingToAction(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	var strike_object *strike.Response_structure

	route := request.User_session_variables.Route
	if route[0] == "Place order again🥚" {
		strike_object = GetOrderData(request)
	} else {
		strike_object = GetHistoryById(request)
	}
	return strike_object
}

//##############################################HELPER##########################################################

func DecodeRequestToStruct(r io.Reader) schema.Strike_Meta_Request_Structure {
	decoder := json.NewDecoder(r)
	var request schema.Strike_Meta_Request_Structure
	err := decoder.Decode(&request)
	if err != nil {
		log.Println("[DECODING ERROR]: ", err)
	}
	log.Println("[REDDEGGS]: ", request)
	return request
}

func WriteResponse(w http.ResponseWriter, s *strike.Response_structure) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(s.ToJson())
}
