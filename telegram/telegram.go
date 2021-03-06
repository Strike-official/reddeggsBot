package telegram

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Strike-official/reddeggsBot/schema"
	"github.com/shashank404error/draco/telegram"
)

func PushToTelegram(apiToken string, request schema.Strike_Meta_Request_Structure, item_description []string, item_total int64, delivery_date, orderID string) {
	name := request.Bybrisk_session_variables.Username
	phone := request.Bybrisk_session_variables.Phone
	address := request.Bybrisk_session_variables.Address

	latitude := fmt.Sprintf("%f", request.Bybrisk_session_variables.Location.Latitude)
	longitude := fmt.Sprintf("%f", request.Bybrisk_session_variables.Location.Longitude)
	mapLink := "https://www.google.com/maps/dir/?api=1&destination=" + latitude + "," + longitude

	var final_item_desc string
	for index, item_description_indi := range item_description {
		if index == len(item_description)-1 {
			final_item_desc = final_item_desc + strings.TrimSpace(item_description_indi)
		} else {
			final_item_desc = final_item_desc + strings.TrimSpace(item_description_indi) + ", "
		}
	}

	quantity := request.User_session_variables.OrderQuantity
	item_total_string := strconv.FormatInt(item_total, 10)

	msg := "ššš° šš«ššš« " + orderID + "\n\nššš¦š: " + name + "\nšš”šØš§š: " + phone + "\nšššš«šš¬š¬: " + address +
		"\n\nšš­šš¦: " + final_item_desc + "\nšš®šš§š­š¢š­š²: " + quantity +
		"\n\nššš„š¢šÆšš«š² ššš­š: " + delivery_date + "\nšš­šš¦ ššØš­šš„: " + item_total_string +
		"\n\n" + mapLink
	telegram.SendNotificationToAllListners(apiToken, msg)
}
