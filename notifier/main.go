package worker

import (
	"log"
	"time"
	"strings"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)


type BotConfig struct {
	ChatID			  int     `json:"chatID"`
	Time    		  string 	 `json:"time"`
	Notification      string     `json:"notification"`
	ImageFileId       string     `json:"imageFileId"`
}


// Worker will do its Action once every interval, making up for lost time that 
// happened during the Action by only waiting the time left in the interval. 
type Worker struct {
	Working         bool          // A flag determining the state of the worker
	ShutdownChannel chan string   // A channel to communicate to the routine
	Interval        time.Duration // The interval with which to run the Action
	period          time.Duration // The actual period of the wait
	Config struct {
		ChatID			  int     `json:"chatID"`
		Time    		  string 	 `json:"time"`
		Notification      string     `json:"notification"`
		ImageFileId       string     `json:"imageFileId"`
	}
}

// NewWorker creates a new worker and instantiates all the data structures required.
func NewWorker(interval time.Duration) *Worker {
	return &Worker{
		Working:         false,
		ShutdownChannel: make(chan string),
		Interval:        interval,
		period:          interval,
		Config:          BotConfig{},
	}
}

// Run starts the worker and listens for a shutdown call.
func (w *Worker) Run(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	w.Working = true
	log.Println("Worker Started")

	// Loop that runs forever
	for {
		select {
		case <-w.ShutdownChannel:
			w.ShutdownChannel <- "Down"
			return
		case <-time.After(w.period):
			// This breaks out of the select, not the for loop.
			break
		}

		started := time.Now()
		w.Action(bot, msg)
		finished := time.Now()

		duration := finished.Sub(started)
		w.period = w.Interval - duration

	}

}

// Shutdown is a graceful shutdown mechanism 
func (w *Worker) Shutdown() {
	w.Working = false

	w.ShutdownChannel <- "Down"
	<-w.ShutdownChannel

	close(w.ShutdownChannel)
}

// Action defines what the worker does; override this. 
func (w *Worker) Action(bot *tgbotapi.BotAPI, msg tgbotapi.MessageConfig) {
	notification := tgbotapi.NewMessage(int64(w.Config.ChatID), w.Config.Notification)

	timesArray := strings.Split(w.Config.Time, ",")

	dt := time.Now()
	var currentTime string = dt.Format("15:04")

	if (contains(timesArray, currentTime)){
		if (w.Config.ImageFileId != ""){
			url, err := bot.GetFileDirectURL(w.Config.ImageFileId)
			if err == nil {
				image := tgbotapi.NewPhotoUpload(int64(w.Config.ChatID), url)
				image.FileID = w.Config.ImageFileId
				image.UseExisting = true
				bot.Send(image)
				sendNotification(notification, bot)
			}else{
				sendNotification(notification, bot)
			}
		}else{
			sendNotification(notification, bot)
		}
	}
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
	   if a == str {
		  return true
	   }
	}
	return false
 }


func sendNotification(msg tgbotapi.Chattable, bot *tgbotapi.BotAPI) {
	bot.Send(msg)
}