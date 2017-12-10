package main

import (
	"github.com/go-telegram-bot-api"
	"google.golang.org/appengine"
	"log"
	"strconv"
	"strings"
	"math"
	"math/rand"
	"fmt"
	"time"
	"net/http"
)
//var db *sql.DB
type Destination struct {
	Location 	tgbotapi.Location
	Name 		string
}

type Question struct{
	Question string
	answerOne string
	answerTwo string
	answerThree string
	answerFour string
	answerCorrect int
}

type Quiz struct{
	ChatID int64
	Points int
	Questions []Question
	NumberOfQuestions int
}

type AnswerData struct{
	QuizIndex 		int		`json:"quiz_index"`
	QuestionIndex	int		`json:"question_index"`
	AnswerIndex		int		`json:"answer_index"`
}

//TODO:All these should be stored on database
var locations []Destination
var Quizes []Quiz
var Questions []Question

func main(){
	appengine.Main()
	bot,err:=tgbotapi.NewBotAPI("504794894:AAEFMJO23cydExR-aZ02SMCSLABbsjcdq-8")
	if err!=nil{
		log.Panic(err)
	}
	chosenMaterial:=-1
	bot.Debug = true
	_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert("https://www.google.com:8443/"+bot.Token, "cert.pem"))
	if err != nil {
		log.Fatal(err)
	}
	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)
	for update:=range updates{
		if update.Message==nil{
			//User pressed the button
			if update.CallbackQuery!=nil{
				bot.Send(OnButtonPressed(update,&chosenMaterial))
				continue
			}
			continue
		}
		//User sent location?
		if update.Message.Location!=nil{
			msg,loc:= ReceiveLocation(update)
			bot.Send(msg)
			bot.Send(loc)
			continue
		}
		//Client chosen material
		if chosenMaterial!=-1{
			bot.Send(ChosenMaterial(update,&chosenMaterial))
			continue
		}
		// Check for the message
		bot.Send(RespondToMessage(update))
	}

}
func init(){
	locations = []Destination{}
	addLocations()

	Questions = []Question{}
	addQuestions()

	Quizes = []Quiz{}
	var err error
	//db, err = sql.Open("sqlite3",":memory")
	if err!=nil{
		log.Fatal(err)
	}
}
func addLocations(){
	chorsu:=Destination{Location:tgbotapi.Location{ Latitude:41.326721,Longitude:69.235122}, Name:"Chorsu"}
	registan:=Destination{Location:tgbotapi.Location{Latitude:39.654684, Longitude:66.975731},Name:"Registan"}
	inha:=Destination{Location:tgbotapi.Location{Latitude:41.338525, Longitude:69.334514},Name:"Inha University in Tashkent"}
	vronica:=Destination{Location:tgbotapi.Location{Latitude:41.321838, Longitude:69.264471},Name:"Vronica"}
	locations = append(locations, chorsu,registan,inha,vronica)
}
func addQuestions(){
	q1:=Question{Question:"What is the capital of Uzbekistan?",answerOne:"London",answerTwo:"Tashkent",answerThree:"Tokyo",answerFour:"Paris",answerCorrect:2}
	q2:=Question{Question:"What is the capital of France?",answerOne:"London",answerTwo:"Tashkent",answerThree:"Tokyo",answerFour:"Paris",answerCorrect:4}
	q3:=Question{Question:"What is the capital of UK?",answerOne:"London",answerTwo:"Tashkent",answerThree:"Tokyo",answerFour:"Paris",answerCorrect:1}
	q4:=Question{Question:"What is the capital of US?",answerOne:"Washington",answerTwo:"Athens",answerThree:"Riga",answerFour:"Rome",answerCorrect:1}
	q5:=Question{Question:"What is the capital of Japan?",answerOne:"London",answerTwo:"Tashkent",answerThree:"Tokyo",answerFour:"Paris",answerCorrect:3}
	q6:=Question{Question:"What is the capital of Germany?",answerOne:"Berlin",answerTwo:"Moscow",answerThree:"Madrid",answerFour:"Kiev",answerCorrect:1}
	q7:=Question{Question:"What is the capital of Russian?",answerOne:"Berlin",answerTwo:"Moscow",answerThree:"Madrid",answerFour:"Kiev",answerCorrect:2}
	q8:=Question{Question:"What is the capital of Spain?",answerOne:"Berlin",answerTwo:"Moscow",answerThree:"Madrid",answerFour:"Kiev",answerCorrect:3}
	q9:=Question{Question:"What is the capital of Ukraine?",answerOne:"Berlin",answerTwo:"Moscow",answerThree:"Madrid",answerFour:"Kiev",answerCorrect:4}
	q10:=Question{Question:"What is the capital of Italy?",answerOne:"Washington",answerTwo:"Athens",answerThree:"Riga",answerFour:"Rome",answerCorrect:4}
	q11:=Question{Question:"What is the capital of Greece?",answerOne:"Washington",answerTwo:"Athens",answerThree:"Riga",answerFour:"Rome",answerCorrect:2}
	Questions=append(Questions,q1,q2,q3,q4,q5,q6,q7,q8,q9,q10,q11)
}
// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

/* Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS!!!!!!
// http://en.wikipedia.org/wiki/Haversine_formula*/
func distance(l1,l2 tgbotapi.Location) float64 {
	lat1:=l1.Latitude
	lon1:=l1.Longitude
	lat2:=l2.Latitude
	lon2:=l2.Longitude
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

func RespondToMessage(update tgbotapi.Update) tgbotapi.MessageConfig{
	switch update.Message.Text {
	case "/start","/help":
		{
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hello! I can do these:")
			kb := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Take Quiz", "take_quiz")),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Find Nearest place to you", "find_place")),
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("Calculate", "calculate")))
			msg.ReplyMarkup = kb
			return msg
		}
	}
	msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"I got u proceeding")
	msg.ReplyToMessageID = update.Message.MessageID
	return msg
}

func ReceiveLocation(update tgbotapi.Update) (one tgbotapi.MessageConfig,two tgbotapi.LocationConfig) {
	min:=distance(locations[0].Location,*update.Message.Location)
	obj:=locations[0]
	for i:=1;i<len(locations);i++{
		temp:=distance(locations[i].Location,*update.Message.Location)
		if temp<min{
			min=temp
			obj=locations[i]
		}
	}

	msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"The nearest location to you is : "+obj.Name)
	loc:=tgbotapi.NewLocation(update.Message.Chat.ID,obj.Location.Latitude,obj.Location.Longitude)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	return msg,loc
}

func ChosenMaterial(update tgbotapi.Update, chosenMaterial *int) tgbotapi.MessageConfig{
	size:=strings.Split(update.Message.Text,",")
	if !(len(size)==2){
		msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"You entered wrong number or in a wrong format plz try again!")
		return  msg
	}
	width,err1:=strconv.ParseFloat(size[0],64)
	if err1!=nil{
		msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"You entered wrong number or in a wrong format plz try again!")
		return  msg
	}
	height,err1:=strconv.ParseFloat(size[1],64)
	if err1!=nil{
		msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"You entered wrong number or in a wrong format plz try again!")
		return msg
	}
	var multiplier int
	switch *chosenMaterial {
	case 1:
		multiplier=50
		break
	case 2:
		multiplier=60
		break
	case 3:
		multiplier=80
		break
	default:
		multiplier=0
	}
	msg:=tgbotapi.NewMessage(update.Message.Chat.ID,"The total amount you need to have is: "+strconv.FormatFloat(width*height/float64(multiplier), 'f', 6, 64)+" grams")
	*chosenMaterial=-1
	return msg
}

func OnButtonPressed(update tgbotapi.Update,chosenMaterial *int) tgbotapi.MessageConfig{
	switch update.CallbackQuery.Data {
	case "calculate":
		msg:=tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,"Choose one of the options: ")
		kb:=tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Шпаклевка(50гр /м2)","1"),
			tgbotapi.NewInlineKeyboardButtonData("Клей(60гр/м2)","2"),
			tgbotapi.NewInlineKeyboardButtonData("Цемент(80гр/м2)","3")))
		msg.ReplyMarkup = kb
		return msg
	case "find_place":
		var menuKeyboard = tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonLocation("I agree")))
		msg :=tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,"Do you agree to send me your location?")
		msg.ReplyMarkup = menuKeyboard
		return msg
	case "1","2","3":
		*chosenMaterial,_=strconv.Atoi(update.CallbackQuery.Data)
		msg:=tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,"Enter width and height. Separate with comma (,):")
		return msg
	case "take_quiz":{
		//TODO:Check if the user is currently taking a quiz
		nQuiz:=Quiz{Points:0,ChatID:update.CallbackQuery.Message.Chat.ID,NumberOfQuestions:0,Questions:[]Question{}}
		Quizes=append(Quizes,nQuiz)
		return GiveQuestion(&nQuiz)
	}
	default:
		//We have got an answer

		msg:=tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID,update.CallbackQuery.Data)
		//TODO:Remove quiz from slice when it is done
		return msg
	}
}

func GiveQuestion(quiz *Quiz) tgbotapi.MessageConfig {
	//TODO:Check for questions and send not given one !!!
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(Questions)+1)
	q := Questions[index]
	msg := tgbotapi.NewMessage(quiz.ChatID, q.Question)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(q.answerOne, "ans_"+fmt.Sprint(quiz.NumberOfQuestions)+"_1")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(q.answerTwo, "ans_"+fmt.Sprint(quiz.NumberOfQuestions)+"_2")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(q.answerThree, "ans_"+fmt.Sprint(quiz.NumberOfQuestions)+"_3")),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(q.answerFour, "ans_"+fmt.Sprint(quiz.NumberOfQuestions)+"_4")))
	return msg
}


