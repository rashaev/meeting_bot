package commands

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Meeting represent the meeting object
type Meeting struct {
	ID        int
	CreatedBy string
	RoomID    int
	StartDate time.Time
	Duration  time.Time
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

// DelRoom function delete a room from database
func DelRoom(db *sql.DB, roomNumber string) int64 {
	result, _ := db.Exec("DELETE FROM room WHERE number = $1", roomNumber)
	rowsAffected, _ := result.RowsAffected()
	return rowsAffected
}
