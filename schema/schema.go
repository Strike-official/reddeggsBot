package schema

import (
	"time"

	"github.com/Strike-official/reddeggsBot/configmanager"
)

var Conf *configmanager.AppConfig

type Strike_Meta_Request_Structure struct {

	// Bybrisk variable from strike bot
	//
	Bybrisk_session_variables Bybrisk_session_variables_struct `json: "bybrisk_session_variables"`

	// Our own variable from previous API
	//
	User_session_variables User_session_variables_struct `json: "user_session_variables"`
}

type Bybrisk_session_variables_struct struct {

	// User ID on Bybrisk
	//
	UserId string `json:"userId"`

	// Our own business Id in Bybrisk
	//
	BusinessId string `json:"businessId"`

	// Handler Name for the API chain
	//
	Handler string `json:"handler"`

	// Current location of the user
	//
	Location GeoLocation_struct `json:"location"`

	// Username of the user
	//
	Username string `json:"username"`

	// Address of the user
	//
	Address string `json:"address"`

	// Phone number of the user
	//
	Phone string `json:"phone"`
}

type GeoLocation_struct struct {
	// Latitude
	//
	Latitude float64 `json:"latitude"`

	// Longitude
	//
	Longitude float64 `json:"longitude"`
}

type User_session_variables_struct struct {
	OrderType []string `json:"order_type,omitempty"`
	//Lat_long GeoLocation_struct `json:"lat_long,omitempty"`
	OrderQuantity      []string           `json:"order_quantity,omitempty"`
	OrderDate          []string           `json:"order_date,omitempty"`
	Route              []string           `json:"route,omitempty"`
	DeliveryAddress    string             `json:"delivery_address"`
	DeliveryLocation   GeoLocation_struct `json:"delivery_location"`
	IsCouponExists     []string           `json:"is_coupon_exists"`
	CouponValue        string             `json:"coupon_value"`
	IsCouponApplied    string             `json:"is_coupon_applied"`
	CouponAppliedValue string             `json:"coupon_applied_value"`
	OrderConfirmation  []string           `json:"order_confirmation"`
}

type Wrapper struct {
	Item_description string    `json:"item_description"`
	Item_total       float64   `json:"item_total"`
	Quantity         string    `json:"quantity"`
	Order_time       time.Time `json:"order_time"`
	Delivery_date    string    `json:"delivery_date"`
}

type InputFromUser struct {
	OrderType        string
	OrderQuantity    string
	OrderDate        string
	DeliveryAddress  string
	DeliveryLocation GeoLocation_struct
}
