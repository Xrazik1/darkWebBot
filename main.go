package main


import (
	"io/ioutil"
	"encoding/json"
	"log"
	"fmt"
	"time"
	"strings"
	"os"
	"io"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var CONFIG Config
var isNotifyWorking bool
const TOKEN = "892282901:AAHRs_-3BA_brHK-_Ui9sGVqdZkVZ35GnBM"


type Config struct {
	ChatID			  int     `json:"chatID"`
	Time    		  string 	 `json:"time"`
	Notification      string     `json:"notification"`
	ImageFileId       string     `json:"imageFileId"`
}


// ADMINS: 430343293, 336963447

func contains(arr []string, str string) bool {
	for _, a := range arr {
	   if a == str {
		  return true
	   }
	}
	return false
 }

func notifier(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig){
	notification := tgbotapi.NewMessage(int64(CONFIG.ChatID), CONFIG.Notification)

	timesArray := strings.Split(CONFIG.Time, ",")

	for x := range time.Tick(time.Minute) {
		dt := time.Now()
		var currentTime string = dt.Format("15:04")

		if (contains(timesArray, currentTime)){
			if (isNotifyWorking){
				if (CONFIG.ImageFileId != ""){
					url, err := bot.GetFileDirectURL(CONFIG.ImageFileId)
					if err == nil {
						image := tgbotapi.NewPhotoUpload(int64(CONFIG.ChatID), url)
						image.FileID = CONFIG.ImageFileId
						image.UseExisting = true
						bot.Send(image)
						sendNotification(notification, bot)
					}else{
						sendNotification(notification, bot)
					}
				}
			}else{
				sendNotification(notification, bot)
			}
		}else{
			break;
		}
		fmt.Println(x)
	}
}

func startNotify(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig){
	isNotifyWorking = true;
	msg.Text = "Бот запустил уведомления"
	bot.Send(msg)
	go notifier(bot, msg)
}

func stopNotify(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig){
	isNotifyWorking = false;
	msg.Text = "Бот остановил уведомления"
	bot.Send(msg)
}

func isImageLoaded()(bool){
	if(CONFIG.ImageFileId != "") {
		return true
	}else{
		return false
	}
}

func isConfigJsonEmpty()(bool){
	byt, err := ioutil.ReadFile("config.json")
	if err != nil || byt == nil {
		return true
	}else{
		config, _ := jsonConverter(byt, Config{}, false)

		if (config.Notification == "" && config.Time == "" && config.ChatID == 0){
			return true
		}
	}
	return false
}

func setConfigToJson(config Config){
	var plugConverter []byte

	_, jsonString := jsonConverter(plugConverter, config, true)

	mydata := []byte(jsonString)

	err := ioutil.WriteFile("config.json", mydata, 0777)
	if err != nil {
		panic(err)
	}
}

/* false - json to struct, true - struct to json */
func jsonConverter(byt []byte, data interface{}, flag bool) (Config, string) {
	var structure Config
	var jsonString string

	if flag == false {

		var config = Config{}

		err := json.Unmarshal(byt, &config)
		if err != nil {
			panic(err)
		}

		structure = config

	} else {
		config, err := json.Marshal(data)

		jsonString = string(config)

		if err != nil {
			panic(err)
		}
	}

	return structure, jsonString
}

func sendNotification(msg tgbotapi.Chattable, bot *tgbotapi.BotAPI) {
	bot.Send(msg)
}

func setConfig(Time string, Notification string, ChatID int, ImageFileId string){
	CONFIG.Time = Time
	CONFIG.Notification = Notification
	CONFIG.ChatID = ChatID
	CONFIG.ImageFileId = ImageFileId

	setConfigToJson(CONFIG)
}

func sendConfig(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig){
	var imageStatus string = "не задана"

	if (isImageLoaded() == true){ imageStatus = "задана" }

	if ((CONFIG.Notification != "") && (CONFIG.Time != "") || (CONFIG.ChatID != 0)){
		msg.Text = "ID отслеживаемого чата: " + strconv.Itoa(CONFIG.ChatID) + " \nОбъявление: " + CONFIG.Notification + " \nВремя отправки: " + CONFIG.Time + "\nКартинка: " + imageStatus
		bot.Send(msg)	
	}else{
		msg.Text = "Кофигурация не задана"
		bot.Send(msg)	
	}

}

func setNotification(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig)(bool){
	msg.Text = "Введите уведомление для отправки"
	bot.Send(msg)
	for update := range updates {
		if update.Message != nil {
			if (update.Message.Text == "exit"){
				break
			}
			
			Notification := strings.TrimSpace(update.Message.Text)

			setConfig(CONFIG.Time, Notification, CONFIG.ChatID, CONFIG.ImageFileId)
			msg.Text = "Параметры успешно установлены."
			bot.Send(msg)
			break;
			
		}
	}
	return false
}

func setTime(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig)(bool){
	msg.Text = "Введите время отправки уведомления в виде: часы и минуты через запятую, Пример: 15:30,20:00,03:00"
	bot.Send(msg)
	for update := range updates {
		if update.Message != nil {
			if (update.Message.Text == "exit"){
				break
			}

			TimesStr := strings.TrimSpace(update.Message.Text)

			setConfig(TimesStr, CONFIG.Notification, CONFIG.ChatID, CONFIG.ImageFileId)
			msg.Text = "Параметры успешно установлены."
			bot.Send(msg)
			break;
			
		}
	}
	return false
}

func saveImage(url string)(bool){
    response, e := http.Get(url)
    if e != nil {
        return false
    }
    defer response.Body.Close()


    file, err := os.Create("images/notificationImage.jpg")
    if err != nil {
        return false
    }
    defer file.Close()

    _, err = io.Copy(file, response.Body)
    if err != nil {
        return false
    }
	
	return true
}

func setNotificationImage(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig)(bool){
	msg.Text = "Отправьте мне картинку, и я добавлю её к объявлению"
	bot.Send(msg)
	for update := range updates {
		if update.Message != nil {
			if(update.Message.Text == "exit"){
				break
			}
			
			if (update.Message.Photo != nil){
				url, err := bot.GetFileDirectURL((*update.Message.Photo)[2].FileID)
				if (err == nil){
					var result bool = saveImage(url)
					if (result == true){
						msg.Text = "Картинка успешно добавлена к уведомлению"
						bot.Send(msg)
						setConfig(CONFIG.Time, CONFIG.Notification, CONFIG.ChatID, (*update.Message.Photo)[2].FileID)
						return false
					}else{
						msg.Text = "Произошла ошибка во время загрузки картинки. Для выхода введите exit."
						bot.Send(msg)
						return true
					}
				}else{
					msg.Text = "Произошла ошибка во время загрузки картинки. Для выхода введите exit."
					bot.Send(msg)
					return true
				}				
			}else{
				msg.Text = "Вы не загрузили картинку. Для выхода введите exit."
				bot.Send(msg)
				return true
			}			
		}
	}
	return false
}

func setChatID(updates tgbotapi.UpdatesChannel, bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig)(bool){
	msg.Text = "Отправьте ID чата вида -1001148956734, который необходимо отслеживать"
	bot.Send(msg)
	for update := range updates {
		if update.Message != nil {
			

			if(update.Message.Text == "exit"){
				break
			}
			
			ChatID, err := strconv.Atoi(update.Message.Text)
			if (err != nil){
				msg.Text = "Вы ввели неверный ID. Для выхода нажмите exit."
				bot.Send(msg)
				return true
			}

			setConfig(CONFIG.Time, CONFIG.Notification, ChatID, CONFIG.ImageFileId)
			msg.Text = "Идентификатор успешно задан"
			bot.Send(msg)

			break;
		}
	}
	return false
}

func loadDataFromJson(){
	byt, err := ioutil.ReadFile("config.json")
	if err != nil || byt == nil {
		mydata := []byte(fmt.Sprintf("{\"chatID\": 0, \"time\": \"\", \"notification\": \"\", \"imageFileId\": \"\"}\n"))

		err := ioutil.WriteFile("config.json", mydata, 0777)
		if err != nil {
			panic(err)
		}
	}else{
		configFromJson, _ := jsonConverter(byt, Config{}, false)

		CONFIG = configFromJson
	}
}



func main() {
	loadDataFromJson()

	bot, err := tgbotapi.NewBotAPI(TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}


		if (update.Message.From.UserName != "mikolasav" && update.Message.From.UserName != "EndlessRat"){
			continue;
		}
		
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "help":
				msg.Text = "Введите \n /setTime для установки времени отправки \n /setNotification для установки уведомления \n /showConfig для просмотра настроек \n /setChatID для установки username отслеживаемого чата \n /setNotificationImage для установки картинки объявления \n /startNotify для запуска уведомлений \n /stopNotify для остановки уведомлений"
				bot.Send(msg)
			case "setNotificationImage":
				var repeat bool = true
				for repeat{
					repeat = setNotificationImage(updates, bot, msg)
				}
			case "setTime":
				var repeat bool = true
				for repeat{
					repeat = setTime(updates, bot, msg)
				}
			case "setNotification":
				var repeat bool = true
				for repeat{
					repeat = setNotification(updates, bot, msg)
				}
			case "showConfig":
				sendConfig(bot, msg)
			case "setChatID":
				var repeat bool = true
				for repeat{
					repeat = setChatID(updates, bot, msg)
				}
			case "startNotify":
				startNotify(bot, msg)
			case "start":
				msg.Text = "Введите \n /setConfig для полной настройки бота \n /showConfig для просмотра настроек \n /setChatID для установки username отслеживаемого чата \n /setNotificationImage для установки картинки объявления \n /startNotify для запуска уведомлений \n /startNotify для остановки уведомлений"
				bot.Send(msg)
			case "stopNotify":
				stopNotify(bot, msg)
			default:
				msg.Text = "Я не понимаю эту команду"
				bot.Send(msg)
			}
		}
	}
}