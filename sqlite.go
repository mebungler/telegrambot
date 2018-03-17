package sqlite

import (
	"github.com/go-telegram-bot-api"
	_ "github.com/go-sqlite3"
	"database/sql"
	"fmt"
	"encoding/json"
	"log"
	"time"
)

//region Constants
const (
	sqlite3Str = "sqlite3"
	memory     = "db.db"
)

//endregion

//region Methods
func New() (Sqlite, error) {
	d, err := sql.Open(sqlite3Str, memory)
	return Sqlite{d}, err
}

func DeleteAll(writable writable, db Sqlite) (sql.Result, error) {
	return db.Exec(fmt.Sprintf(`DELETE FROM %s;`, writable.data().TableName))
}

func DeleteData(writable writable, db Sqlite) (sql.Result, error) {
	return db.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id=?;`,
		writable.data().TableName), writable.data().ChatID)
}

func InsertData(writable writable, db Sqlite) (sql.Result, error) {
	return db.Exec(fmt.Sprintf(`INSERT INTO %s (id,data) VALUES (?,?);`,
		writable.data().TableName), writable.data().ChatID, string(writable.data().Text))
}

func InsertAll(writable writable, db Sqlite) (sql.Result, error) {
	tx, err := db.Begin()
	stmt, err := db.Prepare(
		fmt.Sprintf(`INSERT INTO %s (data) VALUES (?);`,
			writable.data().TableName))
	res, err := stmt.Exec(writable.data().Text)
	err = tx.Commit()
	return res, err
}

func UpdateData(writable writable, db Sqlite) (sql.Result, error) {
	return db.Exec(
		fmt.Sprintf(`UPDATE %s SET  id=?,data=? where id=?;`,
			writable.data().TableName),
		writable.data().ChatID,
		writable.data().Text,
		writable.data().ChatID,
	)
}

func SelectData(writable writable, db Sqlite) (*sql.Rows, error) {
	return db.Query(
		fmt.Sprintf(`SELECT * FROM %s WHERE ID=%d;`,
			writable.data().TableName,
			writable.data().ChatID,
		))
}

func SelectAll(db Sqlite, TableName string) (*sql.Rows, error) {
	return db.Query(fmt.Sprintf(`SELECT * FROM %s;`, TableName))
}

func Select(db Sqlite, writable writable) (*sql.Row) {
	return db.QueryRow(
		fmt.Sprintf(`SELECT * FROM %s WHERE id=%d;`,
			writable.data().TableName,
			writable.data().ChatID))
}

//func SelectNil(db Sqlite, quiz *Quiz, TableName string) (*sql.Row) {
//	str := " WHERE "
//	if len(quiz.QuestionIDs) >= 1 {
//		for _, item := range quiz.QuestionIDs {
//			str += fmt.Sprintf("id<>%d AND ", item)
//		}
//		str = str[:len(str)-4]
//	} else {
//		str = ""
//	}
//	return db.QueryRow(
//		fmt.Sprintf(`SELECT * FROM %s %s;`,
//			TableName,
//			str))
//}

func CreateTable(db Sqlite, TableName string) (sql.Result, error) {
	return db.Exec(
		fmt.Sprintf(`CREATE TABLE %s (id integer NOT NULL PRIMARY KEY, data BLOB);`, TableName))
}

func CreateAITable(db Sqlite, TableName string) (sql.Result, error) {
	return db.Exec(
		fmt.Sprintf(`CREATE TABLE %s (id integer NOT NULL PRIMARY KEY AUTOINCREMENT, data BLOB);`, TableName))
}

//endregion

//region Interfaces
type writable interface {
	data() BotData
}

//endregion

//region Structs
type Sqlite struct {
	*sql.DB
}

type BotData struct {
	ChatID    int64
	TableName string
	Text      []byte
}

type Destination struct {
	ID          int
	Location    tgbotapi.Location
	Name        string
	Description string
}

type Question struct {
	ID            int
	Question      string
	Photo         []byte
	AnswerOne     string
	AnswerTwo     string
	AnswerThree   string
	AnswerFour    string
	AnswerCorrect int
}

type Quiz struct {
	ID        int
	Name      string
	Questions []Question
}

type ECenter struct {
	ID          int
	Name        string
	Photo       []byte
	Description string
}

type QuizReport struct {
	User    User
	Points  int
	Answers []Answer
}

type Answer struct {
	QuestionID     int
	AnswerID       int
	QuestionNumber int
}

type Error struct {
	Error   *error
	Time    time.Time
	IsShown bool
}

type User struct {
	Username       string
	UserID         int
	ChatID         int64
	PhoneNumber    string
	ProfilePicture string
	PhotoFile      tgbotapi.File
	Actions        Action
}

type Action struct {
	QuizReport []QuizReport
}

type City struct {
	ID      int
	Name    string
	Regions []Region
}

type Region struct {
	ID               int
	Name             string
	Shops            []Destination
	Masters          []Master
	EducationCenters []Destination
}

type Message struct {
	ID   int64
	Text string
	User User
}

type Master struct {
	ID int
	Name string
	PhoneNumber string
	Details string
}

//endregion

//region Interface implementations
func (q QuizReport) data() BotData {
	temp, err := json.Marshal(q)
	if err != nil {
		log.Panic(err)
	}
	return BotData{ChatID: int64(q.User.UserID), Text: temp, TableName: "quizReports"}
}

func (q Quiz) data() BotData {
	temp, err := json.Marshal(q)
	if err != nil {
		log.Panic(err)
	}
	return BotData{Text: temp, TableName: "quizes"}
}

func (q Master) data() BotData {
	temp, err := json.Marshal(q)
	if err != nil {
		log.Panic(err)
	}
	return BotData{Text: temp, TableName: "masters"}
}

func (q Question) data() BotData {
	temp, err := json.Marshal(q)
	if err != nil {
		log.Panic(err)
	}
	return BotData{ChatID: int64(q.ID), Text: temp, TableName: "questions"}
}

func (d Destination) data() BotData {
	temp, err := json.Marshal(d)
	if err != nil {
		return BotData{TableName: "error", Text: []byte(err.Error())}
	}
	return BotData{ChatID: int64(d.ID), TableName: "destinations", Text: temp}
}

func (err Error) data() BotData {
	temp, e := json.Marshal(err)
	if e != nil {
		return BotData{TableName: "error", Text: []byte(e.Error())}
	}
	return BotData{TableName: "error", Text: temp}
}

func (u User) data() BotData {
	temp, e := json.Marshal(u)
	if e != nil {
		return BotData{TableName: "error", Text: []byte(e.Error())}
	}
	return BotData{TableName: "users", Text: temp, ChatID: u.ChatID}
}

func (m Message) data() BotData {
	temp, e := json.Marshal(m)
	if e != nil {
		return BotData{TableName: "error", Text: []byte(e.Error())}
	}
	if m.ID != 0 {
		return BotData{TableName: "messages", Text: temp, ChatID: m.ID}
	}
	return BotData{TableName: "messages", Text: temp}
}

func (m City) data() BotData {
	temp, e := json.Marshal(m)
	if e != nil {
		return BotData{TableName: "error", Text: []byte(e.Error())}
	}
	if m.ID != 0 {
		return BotData{TableName: "cities", Text: temp, ChatID: int64(m.ID)}
	}
	return BotData{TableName: "cities", Text: temp}
}

func (m ECenter) data() BotData{
	temp, e := json.Marshal(m)
	if e != nil {
		return BotData{TableName: "error", Text: []byte(e.Error())}
	}
	if m.ID != 0 {
		return BotData{TableName: "ecenters", Text: temp, ChatID: int64(m.ID)}
	}
	return BotData{TableName: "ecenters", Text: temp}
}

//endregion
