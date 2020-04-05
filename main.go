package main

import (
	"database/sql"
	"fmt"
	"meeting_bot/internal/commands"
	"meeting_bot/internal/config"
	logger "meeting_bot/internal/log"
	"net/http"
	"strconv"

	tgcalendar "github.com/dipsycat/calendar-telegram-go"
	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	_ "github.com/jackc/pgx/v4/stdlib"
)

func roomButtons(db *sql.DB) tgbotapi.InlineKeyboardMarkup {
	var keyboardRoom [][]tgbotapi.InlineKeyboardButton
	roomSlice, _ := commands.ListRooms(db)
	for _, room := range roomSlice {
		keyButton := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(room), strconv.Itoa(room)))
		keyboardRoom = append(keyboardRoom, keyButton)
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboardRoom...)
}

func meetingButtons(db *sql.DB, update tgbotapi.Update) tgbotapi.InlineKeyboardMarkup {
	var keyboardMeeting [][]tgbotapi.InlineKeyboardButton
	meetingSlc, _ := commands.GetMyMeetings(db, update)
	for _, meeting := range meetingSlc {
		keyButton := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(strconv.Itoa(meeting.Room)+"\t\t\t"+meeting.StartDate.Format("2006-01-02 15:04:05")+"\t\t\t"+meeting.Duration, strconv.Itoa(meeting.Room)+";"+meeting.StartDate.Format("2006-01-02 15:04:05")+";"+meeting.Duration))
		keyboardMeeting = append(keyboardMeeting, keyButton)
	}
	return tgbotapi.NewInlineKeyboardMarkup(keyboardMeeting...)
}

func callbackHandlerSelectDate(rdb *redis.Client, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	rdb.HSet(strconv.Itoa(update.CallbackQuery.From.ID), "date", update.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Choose time:")
	timeKeyboard := makeTimeKeyboard(update)
	msg.ReplyMarkup = timeKeyboard
	bot.Send(msg)
}

func callbackHandlerSelectDuration(rdb *redis.Client, bot *tgbotapi.BotAPI, update tgbotapi.Update, db *sql.DB) {
	rdb.HSet(strconv.Itoa(update.CallbackQuery.From.ID), "duration", update.CallbackQuery.Data)
	mapResult := rdb.HGetAll(strconv.Itoa(update.CallbackQuery.From.ID)).Val()
	err := commands.AddMeeting(db, update, mapResult)
	if err != nil {
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprint(err))
		bot.Send(msg)
	} else {
		rdb.Del(strconv.Itoa(update.CallbackQuery.From.ID))
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Meeting successfully added")
		bot.Send(msg)
	}
}
func callbackHandlerSelectRoom(rdb *redis.Client, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	rdb.HSet(strconv.Itoa(update.CallbackQuery.From.ID), "room", update.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Select date:")
	genCalendar := tgcalendar.GenerateCalendar(2020, 4)
	msg.ReplyMarkup = genCalendar
	bot.Send(msg)
}

func callbackHandlerSelectTime(rdb *redis.Client, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	var durationKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("30 minute", "30m"),
			tgbotapi.NewInlineKeyboardButtonData("1 hour", "60m"),
			tgbotapi.NewInlineKeyboardButtonData("2 hours", "120m"),
			tgbotapi.NewInlineKeyboardButtonData("3 hours", "180m"),
		),
	)

	rdb.HSet(strconv.Itoa(update.CallbackQuery.From.ID), "time", update.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Choose duration:")
	msg.ReplyMarkup = durationKeyboard
	bot.Send(msg)
}

func makeTimeKeyboard(update tgbotapi.Update) tgbotapi.InlineKeyboardMarkup {
	var timeKeyboard = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("00:00", "00:00"),
			tgbotapi.NewInlineKeyboardButtonData("01:00", "01:00"),
			tgbotapi.NewInlineKeyboardButtonData("02:00", "02:00"),
			tgbotapi.NewInlineKeyboardButtonData("03:00", "03:00"),
			tgbotapi.NewInlineKeyboardButtonData("04:00", "04:00"),
			tgbotapi.NewInlineKeyboardButtonData("05:00", "05:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("06:00", "06:00"),
			tgbotapi.NewInlineKeyboardButtonData("07:00", "07:00"),
			tgbotapi.NewInlineKeyboardButtonData("08:00", "08:00"),
			tgbotapi.NewInlineKeyboardButtonData("09:00", "09:00"),
			tgbotapi.NewInlineKeyboardButtonData("10:00", "10:00"),
			tgbotapi.NewInlineKeyboardButtonData("11:00", "11:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("12:00", "12:00"),
			tgbotapi.NewInlineKeyboardButtonData("13:00", "13:00"),
			tgbotapi.NewInlineKeyboardButtonData("14:00", "14:00"),
			tgbotapi.NewInlineKeyboardButtonData("15:00", "15:00"),
			tgbotapi.NewInlineKeyboardButtonData("16:00", "16:00"),
			tgbotapi.NewInlineKeyboardButtonData("17:00", "17:00"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("18:00", "18:00"),
			tgbotapi.NewInlineKeyboardButtonData("19:00", "19:00"),
			tgbotapi.NewInlineKeyboardButtonData("20:00", "20:00"),
			tgbotapi.NewInlineKeyboardButtonData("21:00", "21:00"),
			tgbotapi.NewInlineKeyboardButtonData("22:00", "22:00"),
			tgbotapi.NewInlineKeyboardButtonData("23:00", "23:00"),
		),
	)
	return timeKeyboard
}

func main() {
	logger := logger.InitLogger("/var/log/meeting_bot/meeting_bot.log", "info")
	cfg, err := config.InitConfig()
	if err != nil {
		logger.Fatal(err)
	}

	dsn := fmt.Sprintf("user=%s dbname=%s password=%s port=%d sslmode=verify­full", cfg.Database.Username, cfg.Database.Name, cfg.Database.Password, cfg.Database.Port)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Fatal("failed to load driver")
	}
	db.Close()

	if err := db.Ping(); err != nil {
		logger.Fatal("Error database connection")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})
	_, err = rdb.Ping().Result()
	if err != nil {
		logger.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Info("Authorized on account ", bot.Self.UserName)
	}

	go http.ListenAndServeTLS(cfg.Network.Host+":"+cfg.Network.Port, cfg.CertFile, cfg.KeyFile, nil)
	logger.Info("Running service  ", cfg.Network.Host, ":", cfg.Network.Port)

	_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert(cfg.Telegram.WebHookURL+bot.Token, cfg.CertFile))
	if err != nil {
		logger.Fatal(err)
	} else {
		logger.Info("Webhook URL: ", cfg.Telegram.WebHookURL+bot.Token)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() == true {
				switch command := update.Message.Command(); command {
				case "delroom":
					go func() {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Which room you want to delete?")
						roomKeyBoard := roomButtons(db)
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
								logger.Info("New room was added")
							} else {
								bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Error adding room"))
								logger.Error(err)
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
					go func() {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Select room:")
						roomKeyBoard := roomButtons(db)
						msg.ReplyMarkup = roomKeyBoard
						bot.Send(msg)
					}()
				case "mymeetings":
					go func() {
						var resultMeetings string
						meetingSlice, err := commands.GetMyMeetings(db, update)
						if err != nil {
							fmt.Println(err)
						}

						for _, meeting := range meetingSlice {
							resultMeetings = resultMeetings + strconv.Itoa(meeting.Room) + "\t\t\t" + meeting.StartDate.Format("2006-01-02 15:04:05") + "\t\t\t" + meeting.Duration + "\n"
						}
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resultMeetings))
					}()
				case "delmeeting":
					go func() {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Which meeting you want to delete?")
						meetingKeyBoard := meetingButtons(db, update)
						msg.ReplyMarkup = meetingKeyBoard
						bot.Send(msg)
					}()
				}
			}
		} else if update.CallbackQuery != nil {
			switch callbackMessage := update.CallbackQuery.Message.Text; callbackMessage {
			case "Which room you want to delete?":
				go func() {
					deletedRows := commands.DelRoom(db, update.CallbackQuery.Data)
					if deletedRows == 0 {
						logger.Warnf("Cannot delete %s from database", update.CallbackQuery.Data)
					} else if deletedRows == 1 {
						bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data+" deleted"))
					}
				}()
			case "Select date:":
				go callbackHandlerSelectDate(rdb, bot, update)
			case "Choose time:":
				go callbackHandlerSelectTime(rdb, bot, update)
			case "Select room:":
				go callbackHandlerSelectRoom(rdb, bot, update)
			case "Choose duration:":
				go callbackHandlerSelectDuration(rdb, bot, update, db)
			case "Which meeting you want to delete?":
				go func() {
					deletedRows := commands.DelMeeting(db, update)
					if deletedRows == 0 {
						logger.Warnf("Cannot delete meeting from database")
					} else if deletedRows == 1 {
						bot.Send(tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Meeting successfully deleted"))
					}
				}()
			}

		}
	}
}
