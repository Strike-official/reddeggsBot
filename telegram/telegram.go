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

	msg := "𝐍𝐞𝐰 𝐎𝐫𝐝𝐞𝐫 " + orderID + "\n\n𝐍𝐚𝐦𝐞: " + name + "\n𝐏𝐡𝐨𝐧𝐞: " + phone + "\n𝐀𝐝𝐝𝐫𝐞𝐬𝐬: " + address +
		"\n\n𝐈𝐭𝐞𝐦: " + final_item_desc + "\n𝐐𝐮𝐚𝐧𝐭𝐢𝐭𝐲: " + quantity +
		"\n\n𝐃𝐞𝐥𝐢𝐯𝐞𝐫𝐲 𝐃𝐚𝐭𝐞: " + delivery_date + "\n𝐈𝐭𝐞𝐦 𝐓𝐨𝐭𝐚𝐥: " + item_total_string +
		"\n\n" + mapLink
	telegram.SendNotificationToAllListners(apiToken, msg)
}
