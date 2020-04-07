package commands

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Meeting represent the meeting object
type Meeting struct {
	CreatedBy int
	Room      int
	StartDate time.Time
	Duration  string
}

//AddRoom function adds new room
func AddRoom(db *sql.DB, roomNum int) error {
	var num int
	row := db.QueryRow("SELECT number FROM room WHERE number = $1", roomNum)

	err := row.Scan(&num)

	if err == sql.ErrNoRows {
		_, err := db.Exec("INSERT INTO room(number) VALUES ($1)", roomNum)
		if err != nil {
			return err
		}
	} else if err == nil {
		err := fmt.Errorf("The room %d already exists", roomNum)
		return err
	}
	return nil
}

// ListRooms function return list of rooms
func ListRooms(db *sql.DB) ([]int, error) {
	var num int
	var result []int

	rows, err := db.Query("SELECT number FROM room")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		if err := rows.Scan(&num); err != nil {
			return nil, err
		}
		result = append(result, num)
	}
	sort.Ints(result)
	return result, err
}

// GetMyMeetings function output user's meetings
func GetMyMeetings(db *sql.DB, update tgbotapi.Update) ([]Meeting, error) {
	var myMeeting Meeting
	var result []Meeting
	rows, err := db.Query("SELECT * FROM meeting WHERE created_by = $1 AND start_date >= $2 ORDER BY start_date", update.Message.From.ID, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&myMeeting.CreatedBy, &myMeeting.Room, &myMeeting.StartDate, &myMeeting.Duration); err != nil {
			return nil, err
		}
		result = append(result, myMeeting)
	}
	return result, err
}

// DelRoom function delete a room from database
func DelRoom(db *sql.DB, roomNumber string) int64 {
	result, _ := db.Exec("DELETE FROM room WHERE number = $1", roomNumber)
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected
}

// DelMeeting function delete a meeting from database
func DelMeeting(db *sql.DB, update tgbotapi.Update) int64 {
	room := strings.Split(update.CallbackQuery.Data, ";")[0]
	datetime := strings.Split(update.CallbackQuery.Data, ";")[1]
	duration := strings.Split(update.CallbackQuery.Data, ";")[2]

	result, _ := db.Exec("DELETE FROM meeting WHERE room = $1 AND start_date = $2 AND duration = $3", room, datetime, duration)
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected
}

// AddMeeting function add a new meeting to database
func AddMeeting(db *sql.DB, update tgbotapi.Update, mapResult map[string]string) error {
	var roomNumber int
	var startDatetime, endDatetime time.Time

	roomParse, _ := strconv.Atoi(mapResult["room"])
	startDateParse, _ := time.Parse("2006.01.2 15:04", mapResult["date"]+" "+mapResult["time"])
	durationParse, _ := mapResult["duration"]
	durationMinutes, _ := time.ParseDuration(durationParse)

	row := db.QueryRow("SELECT room, start_date, start_date + duration as end_date FROM meeting WHERE room = $1 AND start_date < $2 AND start_date + duration > $3", roomParse, startDateParse.Add(durationMinutes), startDateParse)
	err := row.Scan(&roomNumber, &startDatetime, &endDatetime)
	if err == sql.ErrNoRows {
		_, err = db.Exec("INSERT INTO meeting(created_by, room, start_date, duration) VALUES ($1, $2, $3, $4)", update.CallbackQuery.From.ID, roomParse, startDateParse, durationParse)
		if err != nil {
			return err
		}
	} else if err == nil {
		err := fmt.Errorf("This time is already taken. Choose another")
		return err
	}
	return nil
}
