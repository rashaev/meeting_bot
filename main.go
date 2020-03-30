package main

import (
	"database/sql"
	"fmt"
	"meeting_bot/internal/commands"
	"meeting_bot/internal/config"
	"meeting_bot/internal/log"
	"net/http"
	"strconv"

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
						//msg.ReplyMarkup = numericKeyboard
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
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Add meeting:")
					genCalendar := tgcalendar.GenerateCalendar(2020, 3)
					msg.ReplyMarkup = genCalendar
					bot.Send(msg)
				default:
					fmt.Println("command not found")
				}
			}
		} else if update.CallbackQuery != nil {
			switch callbackMassage := update.CallbackQuery.Message.Text; callbackMassage {
			case "Which room you want to delete?":
				go func() {
					deletedRows := commands.DelRoom(db, update.CallbackQuery.Data)
					if deletedRows == 0 {
						log.Warnf("Cannot delete %s from database", update.CallbackQuery.Data)
					} else if deletedRows == 1 {
						bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data+" deleted"))
					}
				}()
			case "Add meeting:":
				fmt.Println(update.CallbackQuery.Data)
			}

		}
	}
}
