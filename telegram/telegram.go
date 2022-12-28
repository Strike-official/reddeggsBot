package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Strike-official/reddeggsBot/configmanager"
	"github.com/Strike-official/reddeggsBot/schema"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func PushToTelegram(apiToken string, request schema.Strike_Meta_Request_Structure, item_description []string, item_total int64, delivery_date, orderID string) {
	schema.Conf = configmanager.GetAppConfig()
	name := request.Bybrisk_session_variables.Username
	phone := request.Bybrisk_session_variables.Phone
	address := request.User_session_variables.DeliveryAddress

	latitude := fmt.Sprintf("%f", request.User_session_variables.DeliveryLocation.Latitude)
	longitude := fmt.Sprintf("%f", request.User_session_variables.DeliveryLocation.Longitude)
	mapLink := "https://www.google.com/maps/dir/?api=1&destination=" + latitude + "," + longitude

	var final_item_desc string
	for index, item_description_indi := range item_description {
		if index == len(item_description)-1 {
			final_item_desc = final_item_desc + strings.TrimSpace(item_description_indi)
		} else {
			final_item_desc = final_item_desc + strings.TrimSpace(item_description_indi) + ", "
		}
	}

	quantity := request.User_session_variables.OrderQuantity[0]
	item_total_string := strconv.FormatInt(item_total, 10)

	msg := "𝐍𝐞𝐰 𝐎𝐫𝐝𝐞𝐫 " + orderID + "\n\n𝐍𝐚𝐦𝐞: " + name + "\n𝐏𝐡𝐨𝐧𝐞: " + phone + "\n𝐀𝐝𝐝𝐫𝐞𝐬𝐬: " + address +
		"\n\n𝐈𝐭𝐞𝐦: " + final_item_desc + "\n𝐐𝐮𝐚𝐧𝐭𝐢𝐭𝐲: " + quantity +
		"\n\n𝐃𝐞𝐥𝐢𝐯𝐞𝐫𝐲 𝐃𝐚𝐭𝐞: " + delivery_date + "\n𝐈𝐭𝐞𝐦 𝐓𝐨𝐭𝐚𝐥: " + item_total_string +
		"\n\n" + mapLink
	sendNotificationToAllListners(apiToken, msg, schema.Conf.TelegramChatID)
}

func sendNotificationToAllListners(apiToken string, body string, chatID string) {
	bot, _ := newBot(apiToken, false)

	log.Printf("[TELEGRAM PUSH] bot: %s body: %s", bot.Self.UserName, body)
	i, _ := strconv.ParseInt(chatID, 10, 64)
	msg := tgbotapi.NewMessage(i, body)
	bot.Send(msg)
}
func newBot(token string, debug bool) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Println("[Telegram Bot Error] Failed to get new bot: ", err)
		return bot, err
	}
	bot.Debug = debug
	return bot, err
}
