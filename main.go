package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/strike-official/go-sdk/strike"

	"strconv"

	"github.com/Strike-official/reddeggsBot/configmanager"
	"github.com/Strike-official/reddeggsBot/database"
	"github.com/Strike-official/reddeggsBot/schema"
	"github.com/Strike-official/reddeggsBot/telegram"
)

var (
	base_url             = ""
	TelegramAPIToken     = ""
	ActiveCouponCode     = "strike10"
	GotDelayedCouponCode = "weareimproving20"
)

func determineListenAddress(port string) (string, error) {
	if port == "" {
		port = "6005"
	}
	return ":" + port, nil
}

func main() {
	err := configmanager.InitAppConfig("configs/config.json")
	if err != nil {
		log.Fatal("[startAPIs] Failed to start APIs. Error: ", err)
	}
	schema.Conf = configmanager.GetAppConfig()

	base_url = schema.Conf.APIEp
	TelegramAPIToken = schema.Conf.TelegramAPIToken

	addr, err := determineListenAddress(schema.Conf.Port)
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	//r.HandleFunc("/",index).Methods("POST")
	r.HandleFunc("/reddeggs/order-view", order_view).Methods("POST")
	r.HandleFunc("/reddeggs/coupon/show/option", coupon_show_option).Methods("POST")
	r.HandleFunc("/reddeggs/coupon/get/input", coupon_get_input).Methods("POST")
	r.HandleFunc("/reddeggs/coupon/process", coupon_process).Methods("POST")
	r.HandleFunc("/reddeggs/order-confirmation-input", order_confirmation_input).Methods("POST")
	r.HandleFunc("/reddeggs/order-controller-process", order_confirmation_process).Methods("POST")
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

var (
	price30EggsTray                        int64 = 200
	price12PackEggs                        int64 = 100
	defaultPrice30EggsTray                 int64 = 180
	defaultPrice12PackEggs                 int64 = 90
	originalPriceWithoutDiscount30EggsTray int64 = 200
	originalPriceWithoutDiscount12PackEggs int64 = 100
	discountedPrice30EggsTray              int64 = 180 // 10% of the price
	discountedPrice12PackEggs              int64 = 90
	discountedPriceForDelay30EggsTray      int64 = 160 // 20% of the price
	discountedPriceForDelay12PackEggs      int64 = 80
)
var product_name_pic map[string]productDesc

// #####################################################HANDLERS###########################################################
func order_view(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, GetOrderData(DecodeRequestToStruct(r.Body)))
}

func coupon_show_option(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, CouponShowOption(DecodeRequestToStruct(r.Body)))
}

func coupon_get_input(w http.ResponseWriter, r *http.Request) {

	latitude, _ := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	longitude, _ := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	in := schema.InputFromUser{
		OrderType:       r.URL.Query().Get("order_type"),
		OrderQuantity:   r.URL.Query().Get("order_quantity"),
		OrderDate:       r.URL.Query().Get("order_date"),
		DeliveryAddress: r.URL.Query().Get("delivery_address"),
		DeliveryLocation: schema.GeoLocation_struct{
			Latitude:  latitude,
			Longitude: longitude,
		},
	}

	WriteResponse(w, CouponGetInput(DecodeRequestToStruct(r.Body), in))
}

func coupon_process(w http.ResponseWriter, r *http.Request) {
	latitude, _ := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	longitude, _ := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	in := schema.InputFromUser{
		OrderType:       r.URL.Query().Get("order_type"),
		OrderQuantity:   r.URL.Query().Get("order_quantity"),
		OrderDate:       r.URL.Query().Get("order_date"),
		DeliveryAddress: r.URL.Query().Get("delivery_address"),
		DeliveryLocation: schema.GeoLocation_struct{
			Latitude:  latitude,
			Longitude: longitude,
		},
	}
	WriteResponse(w, CouponProcess(DecodeRequestToStruct(r.Body), in))
}

func order_confirmation_input(w http.ResponseWriter, r *http.Request) {
	WriteResponse(w, OrderConfirmationInput(DecodeRequestToStruct(r.Body)))
}

func order_confirmation_process(w http.ResponseWriter, r *http.Request) {
	latitude, _ := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	longitude, _ := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	in := schema.InputFromUser{
		OrderType:       r.URL.Query().Get("order_type"),
		OrderQuantity:   r.URL.Query().Get("order_quantity"),
		OrderDate:       r.URL.Query().Get("order_date"),
		DeliveryAddress: r.URL.Query().Get("delivery_address"),
		DeliveryLocation: schema.GeoLocation_struct{
			Latitude:  latitude,
			Longitude: longitude,
		},
	}
	WriteResponse(w, OrderConfirmationProcess(DecodeRequestToStruct(r.Body), in))
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
	price30EggsTray = defaultPrice30EggsTray
	price12PackEggs = defaultPrice12PackEggs
	product_name_pic = map[string]productDesc{
		"    30 Eggs Tray": {
			PicURL: "http://reddeggs.com/Product_Images/600f2139-b74e-4ba9-bdaf-4944d5b9808c.jpg",
			Price:  originalPriceWithoutDiscount30EggsTray,
		},
		"    12 Pack Eggs": {
			PicURL: "http://reddeggs.com/Product_Images/0c6e5ed4-6e8e-40f2-8f14-ce8c58e74c66.jpeg",
			Price:  originalPriceWithoutDiscount12PackEggs,
		},
	}

	name := request.Bybrisk_session_variables.Username
	strike_object := strike.Create("order-controller", base_url+"/coupon/show/option")

	question_object1 := strike_object.Question("order_type").
		QuestionText().SetTextToQuestion("Namaste "+name+" 🙏 select a product from below. Press long for multiple selection", "Text Description, getting used for testing purpose.")

	answer_object1 := question_object1.Answer(true).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)

	answer_object1.AnswerCard().SetHeaderToAnswer(2, "WRAP").
		AddGraphicRowToAnswer(strike.PICTURE_ROW, []string{"http://reddeggs.com/Product_Images/600f2139-b74e-4ba9-bdaf-4944d5b9808c.jpg"}, []string{}).
		AddTextRowToAnswer(strike.H4, "    30 Eggs Tray", "#3c6dbd", true).
		AddTextRowToAnswer(strike.H4, "    ₹"+strconv.FormatInt(price30EggsTray, 10), "#009646", true).
		AddTextRowToAnswer(strike.H4, "    ₹̶2̶0̶0̶", "#ed0212", true).
		AddTextRowToAnswer(strike.H6, "¹ Upon using\ncoupon code STRIKE10", "#009646", false).
		AnswerCard().SetHeaderToAnswer(2, "WRAP").
		AddGraphicRowToAnswer(strike.PICTURE_ROW, []string{"http://reddeggs.com/Product_Images/0c6e5ed4-6e8e-40f2-8f14-ce8c58e74c66.jpeg"}, []string{}).
		AddTextRowToAnswer(strike.H4, "    12 Pack Eggs", "#3c6dbd", true).
		AddTextRowToAnswer(strike.H4, "    ₹"+strconv.FormatInt(price12PackEggs, 10), "#009646", true).
		AddTextRowToAnswer(strike.H4, "    ₹̶1̶0̶0̶", "#ed0212", true).
		AddTextRowToAnswer(strike.H6, "¹ Upon using\ncoupon code STRIKE10", "#009646", false)

	question_object2 := strike_object.Question("order_quantity").
		QuestionText().SetTextToQuestion("Great! how many trays do you want?", "Text Description, getting used for testing purpose.")

	answer_object2 := question_object2.Answer(false).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)
	answer_object2.AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "1", "#3c6dbd", true).
		AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "2", "#3c6dbd", true).
		AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "3", "#3c6dbd", true).
		AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "4", "#3c6dbd", true).
		AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "5", "#3c6dbd", true)

	question_object3 := strike_object.Question("order_date").
		QuestionText().SetTextToQuestion("When should we deliver?", "Text Description, getting used for testing purpose.")

	question_object3.Answer(false).DateInput("Get delivery date")

	question_object4 := strike_object.Question("delivery_address").
		QuestionText().SetTextToQuestion("Please enter the delivery address", "Text Description, getting used for testing purpose.")

	question_object4.Answer(false).TextInput("Address")

	question_object5 := strike_object.Question("delivery_location").
		QuestionText().SetTextToQuestion("Please share your location for a smooth delivery", "Text Description, getting used for testing purpose.")

	question_object5.Answer(false).LocationInput("📍 Select location here")

	return strike_object
}

func CouponShowOption(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	orderType := request.User_session_variables.OrderType             //order_type
	orderQuantity := request.User_session_variables.OrderQuantity     //order_quantity
	orderDate := request.User_session_variables.OrderDate[0]          //order_date
	deliveryAddress := request.User_session_variables.DeliveryAddress //delivery_address
	latitude := fmt.Sprintf("%f", request.User_session_variables.DeliveryLocation.Latitude)
	longitude := fmt.Sprintf("%f", request.User_session_variables.DeliveryLocation.Longitude) //delivery_location

	strike_object := strike.Create("coupon-is-exist", base_url+"/coupon/get/input?order_type="+arrToString(orderType)+"&order_quantity="+arrToString(orderQuantity)+"&order_date="+orderDate+"&delivery_address="+deliveryAddress+"&latitude="+latitude+"&longitude="+longitude)
	question_object1 := strike_object.Question("is_coupon_exists").
		QuestionText().SetTextToQuestion("🏷️ Do you have a coupon code?", "Text Description, getting used for testing purpose.")

	answer_object1 := question_object1.Answer(true).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)
	answer_object1.AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "Yes", "#3c6dbd", true).
		AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "No", "#3c6dbd", true)
	return strike_object
}

func CouponGetInput(request schema.Strike_Meta_Request_Structure, in schema.InputFromUser) *strike.Response_structure {
	isCouponExists := request.User_session_variables.IsCouponExists[0]
	if isCouponExists == "Yes" {
		strike_object := strike.Create("coupon-get-input", base_url+"/coupon/process?order_type="+in.OrderType+"&order_quantity="+in.OrderQuantity+"&order_date="+in.OrderDate+"&delivery_address="+in.DeliveryAddress+"&latitude="+fmt.Sprintf("%f", in.DeliveryLocation.Latitude)+"&longitude="+fmt.Sprintf("%f", in.DeliveryLocation.Longitude))

		question_object1 := strike_object.Question("coupon_value").
			QuestionText().SetTextToQuestion("👇 Enter coupon code below", "Text Description, getting used for testing purpose.")

		question_object1.Answer(false).TextInput("coupon code")
		return strike_object
	} else {
		request.User_session_variables.OrderType = stringToArr(in.OrderType)
		request.User_session_variables.OrderQuantity = stringToArr(in.OrderQuantity)
		request.User_session_variables.OrderDate = []string{in.OrderDate}
		request.User_session_variables.DeliveryAddress = in.DeliveryAddress
		request.User_session_variables.DeliveryLocation = schema.GeoLocation_struct{
			Latitude:  in.DeliveryLocation.Latitude,
			Longitude: in.DeliveryLocation.Longitude,
		}
		request.User_session_variables.IsCouponApplied = "FALSE"
		return OrderConfirmationInput(request)
	}
}

func CouponProcess(request schema.Strike_Meta_Request_Structure, in schema.InputFromUser) *strike.Response_structure {
	userID := request.Bybrisk_session_variables.UserId
	couponValue := request.User_session_variables.CouponValue
	couponValueLowerCase := strings.ToLower(couponValue)
	if couponValueLowerCase == ActiveCouponCode || (couponValueLowerCase == GotDelayedCouponCode && (userID == "63a59481040e6322f17cbbd0" || userID == "63a6faa0040e6322f17cbbd6" || userID == "623efb9195ba637fe92fb07b")) {
		var price30EggsTray int64
		var price12PackEggs int64
		if couponValueLowerCase == ActiveCouponCode {
			price30EggsTray = discountedPrice30EggsTray
			price12PackEggs = discountedPrice12PackEggs
			request.User_session_variables.CouponAppliedValue = ActiveCouponCode
		}
		if couponValueLowerCase == GotDelayedCouponCode {
			price30EggsTray = discountedPriceForDelay30EggsTray
			price12PackEggs = discountedPriceForDelay12PackEggs
			request.User_session_variables.CouponAppliedValue = GotDelayedCouponCode
		}
		product_name_pic = map[string]productDesc{
			"    30 Eggs Tray": {
				PicURL: "http://reddeggs.com/Product_Images/600f2139-b74e-4ba9-bdaf-4944d5b9808c.jpg",
				Price:  price30EggsTray,
			},
			"    12 Pack Eggs": {
				PicURL: "http://reddeggs.com/Product_Images/0c6e5ed4-6e8e-40f2-8f14-ce8c58e74c66.jpeg",
				Price:  price12PackEggs,
			},
		}
		request.User_session_variables.OrderType = stringToArr(in.OrderType)
		request.User_session_variables.OrderQuantity = stringToArr(in.OrderQuantity)
		request.User_session_variables.OrderDate = []string{in.OrderDate}
		request.User_session_variables.DeliveryAddress = in.DeliveryAddress
		request.User_session_variables.DeliveryLocation = schema.GeoLocation_struct{
			Latitude:  in.DeliveryLocation.Latitude,
			Longitude: in.DeliveryLocation.Longitude,
		}
		request.User_session_variables.IsCouponApplied = "TRUE"
		return OrderConfirmationInput(request)
	} else {
		strike_object := strike.Create("coupon-is-exist", base_url+"/coupon/get/input?order_type="+in.OrderType+"&order_quantity="+in.OrderQuantity+"&order_date="+in.OrderDate+"&delivery_address="+in.DeliveryAddress+"&latitude="+fmt.Sprintf("%f", in.DeliveryLocation.Latitude)+"&longitude="+fmt.Sprintf("%f", in.DeliveryLocation.Longitude))
		question_object1 := strike_object.Question("is_coupon_exists").
			QuestionText().SetTextToQuestion("😬 Oops! that code didn't worked. Do you have some other coupon code?", "Text Description, getting used for testing purpose.")

		answer_object1 := question_object1.Answer(false).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)
		answer_object1.AnswerCard().SetHeaderToAnswer(1, "WRAP").
			AddTextRowToAnswer(strike.H4, "Yes", "#3c6dbd", true).
			AnswerCard().SetHeaderToAnswer(1, "WRAP").
			AddTextRowToAnswer(strike.H4, "No", "#3c6dbd", true)
		return strike_object
	}

}

func OrderConfirmationInput(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	orderType := request.User_session_variables.OrderType             //order_type
	orderQuantity := request.User_session_variables.OrderQuantity     //order_quantity
	orderDate := request.User_session_variables.OrderDate[0]          //order_date
	deliveryAddress := request.User_session_variables.DeliveryAddress //delivery_address
	latitude := fmt.Sprintf("%f", request.User_session_variables.DeliveryLocation.Latitude)
	longitude := fmt.Sprintf("%f", request.User_session_variables.DeliveryLocation.Longitude) //delivery_location
	isCouponApplied := request.User_session_variables.IsCouponApplied

	strike_object := strike.Create("coupon-is-exist", base_url+"/order-controller-process?order_type="+arrToString(orderType)+"&order_quantity="+arrToString(orderQuantity)+"&order_date="+orderDate+"&delivery_address="+deliveryAddress+"&latitude="+latitude+"&longitude="+longitude+"&is_coupon_applied="+isCouponApplied)

	var description string
	var total int64
	for i, product := range orderType {
		total = total + product_name_pic[product].Price
		if i == 0 {
			description = product + " (₹ " + strconv.FormatInt(product_name_pic[product].Price, 10) + ")"
			continue
		}
		description = description + "\n" + product + " (₹ " + strconv.FormatInt(product_name_pic[product].Price, 10) + ")"
	}
	quantity_int64, _ := strconv.ParseInt(orderQuantity[0], 10, 64)
	total = total * quantity_int64

	question_object1 := strike_object.Question("order_confirmation").
		QuestionCard().SetHeaderToQuestion(1, strike.HALF_WIDTH)
	if isCouponApplied == "TRUE" {
		couponAppliedValue := request.User_session_variables.CouponAppliedValue
		var couponText string
		if couponAppliedValue == GotDelayedCouponCode {
			couponText = "😃 Yay, we're glad you came back! your 20% offer is applied"
		}
		if couponAppliedValue == ActiveCouponCode {
			couponText = "🎉 offer applied, 10% limited discount"
		}
		question_object1 = question_object1.AddTextRowToQuestion(strike.H4, couponText, "#009646", true)
	}

	question_object1 = question_object1.AddTextRowToQuestion(strike.H4, "👇 Below is your order summary", "black", true).
		AddTextRowToQuestion(strike.H4, description, "black", true).
		AddTextRowToQuestion(strike.H4, "Quantity: "+orderQuantity[0], "#009646", true).
		AddTextRowToQuestion(strike.H4, "Total to be paid: ₹ "+strconv.FormatInt(total, 10), "#009646", true).
		AddTextRowToQuestion(strike.H5, "Deliver to: "+deliveryAddress, "#4e4f52", true).
		AddTextRowToQuestion(strike.H5, "Deliver on: "+orderDate, "#4e4f52", true)

	answer_object1 := question_object1.Answer(false).AnswerCardArray(strike.HORIZONTAL_ORIENTATION)
	answer_object1.AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "🛒 Place order", "#3c6dbd", true).
		AnswerCard().SetHeaderToAnswer(1, "WRAP").
		AddTextRowToAnswer(strike.H4, "Cancel", "#f02618", true)
	return strike_object
}

func OrderConfirmationProcess(request schema.Strike_Meta_Request_Structure, in schema.InputFromUser) *strike.Response_structure {
	confirmationResp := request.User_session_variables.OrderConfirmation[0]
	if confirmationResp == "🛒 Place order" {
		request.User_session_variables.OrderType = stringToArr(in.OrderType)
		request.User_session_variables.OrderQuantity = stringToArr(in.OrderQuantity)
		request.User_session_variables.OrderDate = []string{in.OrderDate}
		request.User_session_variables.DeliveryAddress = in.DeliveryAddress
		request.User_session_variables.DeliveryLocation = schema.GeoLocation_struct{
			Latitude:  in.DeliveryLocation.Latitude,
			Longitude: in.DeliveryLocation.Longitude,
		}
		return PlaceOrder(request)
	} else {
		return GetOrderData(request)
	}
}

func PlaceOrder(request schema.Strike_Meta_Request_Structure) *strike.Response_structure {
	var totalPrice int64
	for _, product := range request.User_session_variables.OrderType {
		totalPrice = totalPrice + product_name_pic[product].Price
	}
	quantity_int64, err := strconv.ParseInt(request.User_session_variables.OrderQuantity[0], 10, 64)
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
			AddTextRowToAnswer(strike.H4, "Place order again!", "#3c6dbd", true)
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
			AddTextRowToQuestion(strike.H4, "Pay ₹"+strconv.FormatInt(totalPrice, 10)+" at the time of delivery", "green", true)

		answer_object1 := question_object1.Answer(false).AnswerCardArray(strike.VERTICAL_ORIENTATION)

		for _, product := range request.User_session_variables.OrderType {
			answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, "WRAP").
				AddGraphicRowToAnswer(strike.PICTURE_ROW, []string{product_name_pic[product].PicURL}, []string{}).
				AddTextRowToAnswer(strike.H4, product, "black", true).
				AddTextRowToAnswer(strike.H4, "₹"+strconv.FormatInt(product_name_pic[product].Price, 10)+" * "+request.User_session_variables.OrderQuantity[0], "#009646", true)

		}

		answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "Expect delivery between 9AM - 7PM "+request.User_session_variables.OrderDate[0], "black", true).
			AddTextRowToAnswer(strike.H4, "Call on +91 9993866695 to modify time", "black", false)

		answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "Place order again🥚", "#3c6dbd", true)
		answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
			AddTextRowToAnswer(strike.H4, "View order history", "#3c6dbd", true)
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
		QuestionCard().SetHeaderToQuestion(1, strike.HALF_WIDTH).AddTextRowToQuestion(strike.H4, "Hi "+name+", below is your order history 👇", "#0088e3", true)

	answer_object1 := question_object1.Answer(false).AnswerCardArray(strike.VERTICAL_ORIENTATION)

	if len(orders) == 0 {
		answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, "WRAP").
			AddTextRowToAnswer(strike.H4, "No orders yet!", "black", false)
	} else {
		for _, order := range orders {
			answer_object1 = answer_object1.AnswerCard().SetHeaderToAnswer(10, "WRAP").
				AddTextRowToAnswer(strike.H4, order.Quantity+" × "+order.Item_description, "#474747", true).
				AddTextRowToAnswer(strike.H4, "₹ "+fmt.Sprintf("%v", order.Item_total), "#009646", true).
				AddTextRowToAnswer(strike.H5, "Ordered on: "+order.Order_time.Format("02 Jan 2006"), "black", true)

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
		AddTextRowToAnswer(strike.H4, "Place order again🥚", "#3c6dbd", true)
	answer_object1.AnswerCard().SetHeaderToAnswer(1, strike.HALF_WIDTH).
		AddTextRowToAnswer(strike.H4, "View order history", "#3c6dbd", true)
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

func initLogger(filePath string) *os.File {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	return file
}

func arrToString(arr []string) string {
	var resp string
	for i, el := range arr {
		if i == 0 {
			resp = el
			continue
		}
		resp = resp + " | " + el
	}
	return resp
}

func stringToArr(s string) []string {
	resp := strings.Split(s, " | ")
	return resp
}
