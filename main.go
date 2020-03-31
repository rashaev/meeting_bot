package main

import (
	"database/sql"
	"fmt"
	"meeting_bot/internal/commands"
	"meeting_bot/internal/config"
	"meeting_bot/internal/log"
	"net/http"
	"strconv"
	"time"

	tgcalendar "github.com/dipsycat/calendar-telegram-go"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func makeButtons(db *sql.DB) tgbotapi.InlineKeyboardMarkup {
	var keyboardRoom [][]tgbotapi.InlineKeyboardButton
	roomSlice, _ := commands.ListRooms(db)
	for _, room := range roomSlice {
		keyButton := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(room), strconv.Itoa(room)))
		keyboardRoom = append(keyboardRoom, keyButton)
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboardRoom...)
}

func makeTimeKeyboard(update tgbotapi.Update) tgbotapi.InlineKeyboardMarkup {
	var timeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("00:00", update.CallbackQuery.Data+" "+"00:00"),
			tgbotapi.NewInlineKeyboardButtonData("01:00", update.CallbackQuery.Data+" "+"01:00"),
			tgbotapi.NewInlineKeyboardButtonData("02:00", update.CallbackQuery.Data+" "+"02:00"),
			tgbotapi.NewInlineKeyboardButtonData("03:00", update.CallbackQuery.Data+" "+"03:00"),
			tgbotapi.NewInlineKeyboardButtonData("04:00", update.CallbackQuery.Data+" "+"04:00"),
			tgbotapi.NewInlineKeyboardButtonData("05:00", update.CallbackQuery.Data+" "+"05:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("06:00", update.CallbackQuery.Data+" "+"06:00"),
			tgbotapi.NewInlineKeyboardButtonData("07:00", update.CallbackQuery.Data+" "+"07:00"),
			tgbotapi.NewInlineKeyboardButtonData("08:00", update.CallbackQuery.Data+" "+"08:00"),
			tgbotapi.NewInlineKeyboardButtonData("09:00", update.CallbackQuery.Data+" "+"09:00"),
			tgbotapi.NewInlineKeyboardButtonData("10:00", update.CallbackQuery.Data+" "+"10:00"),
			tgbotapi.NewInlineKeyboardButtonData("11:00", update.CallbackQuery.Data+" "+"11:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("12:00", update.CallbackQuery.Data+" "+"12:00"),
			tgbotapi.NewInlineKeyboardButtonData("13:00", update.CallbackQuery.Data+" "+"13:00"),
			tgbotapi.NewInlineKeyboardButtonData("14:00", update.CallbackQuery.Data+" "+"14:00"),
			tgbotapi.NewInlineKeyboardButtonData("15:00", update.CallbackQuery.Data+" "+"15:00"),
			tgbotapi.NewInlineKeyboardButtonData("16:00", update.CallbackQuery.Data+" "+"16:00"),
			tgbotapi.NewInlineKeyboardButtonData("17:00", update.CallbackQuery.Data+" "+"17:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("18:00", update.CallbackQuery.Data+" "+"18:00"),
			tgbotapi.NewInlineKeyboardButtonData("19:00", update.CallbackQuery.Data+" "+"19:00"),
			tgbotapi.NewInlineKeyboardButtonData("20:00", update.CallbackQuery.Data+" "+"20:00"),
			tgbotapi.NewInlineKeyboardButtonData("21:00", update.CallbackQuery.Data+" "+"21:00"),
			tgbotapi.NewInlineKeyboardButtonData("22:00", update.CallbackQuery.Data+" "+"22:00"),
			tgbotapi.NewInlineKeyboardButtonData("23:00", update.CallbackQuery.Data+" "+"23:00"),
		),
	)
	return timeKeyboard
}

func main() {
	log := log.InitLogger("meetingbot.log", "info")
	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf("user=%s dbname=%s password=%s port=%d sslmode=verify­full", cfg.Database.Username, cfg.Database.Name, cfg.Database.Password, cfg.Database.Port)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("failed to load driver")
	}

	if err := db.Ping(); err != nil {
		log.Fatal("Error database connection")
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Authorized on account ", bot.Self.UserName)
	}

	go http.ListenAndServeTLS(cfg.Network.Host+":"+cfg.Network.Port, cfg.CertFile, cfg.KeyFile, nil)
	log.Info("Running service  ", cfg.Network.Host, ":", cfg.Network.Port)

	_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert(cfg.Telegram.WebHookURL+bot.Token, cfg.CertFile))
	if err != nil {
		log.Fatal(err)
	} else {
		log.Info("Webhook URL: ", cfg.Telegram.WebHookURL+bot.Token)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() == true {
				switch command := update.Message.Command(); command {
				case "delroom":
					go func() {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Which room you want to delete?")
						roomKeyBoard := makeButtons(db)
						msg.ReplyMarkup = roomKeyBoard
						bot.Send(msg)
					}()
				case "addroom":
					go func() {
						args := update.Message.CommandArguments()
						if roomInt, err := strconv.Atoi(args); err != nil {
							bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "The room should be integer"))
						} else {
							err := commands.AddRoom(db, roomInt)
							if err == nil {
								bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "New room was added: "))
								log.Info("New room was added")
							} else {
								bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error adding room"))
								log.Error(err)
							}

						}
					}()
				case "listrooms":
					go func() {
						var resultRooms string
						roomSlice, _ := commands.ListRooms(db)
						for _, room := range roomSlice {
							resultRooms = resultRooms + strconv.Itoa(room) + "\n"
						}
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resultRooms))
					}()
				case "addmeeting":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select date:")
					genCalendar := tgcalendar.GenerateCalendar(2020, 4)
					msg.ReplyMarkup = genCalendar
					bot.Send(msg)
				default:
					fmt.Println("command not found")
				}
			}
		} else if update.CallbackQuery != nil {
			switch callbackMessage := update.CallbackQuery.Message.Text; callbackMessage {
			case "Which room you want to delete?":
				go func() {
					deletedRows := commands.DelRoom(db, update.CallbackQuery.Data)
					if deletedRows == 0 {
						log.Warnf("Cannot delete %s from database", update.CallbackQuery.Data)
					} else if deletedRows == 1 {
						bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data+" deleted"))
					}
				}()
			case "Select date:":
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Choose time:")
				timeKeyboard := makeTimeKeyboard(update)
				msg.ReplyMarkup = timeKeyboard
				bot.Send(msg)
			case "Choose time:":
				dateTime, _ := time.Parse("2006.01.2 15:04", update.CallbackQuery.Data)
				fmt.Println(dateTime)
			}

		}
	}
}
