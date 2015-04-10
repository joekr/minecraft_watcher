package main

import (
	"flag"
	"github.com/howeyc/fsnotify"
	"github.com/mitsuse/pushbullet-go"
	"github.com/mitsuse/pushbullet-go/requests"
	"log"
	"os"
	"strings"
	"time"
)

var token string
var userName string

func main() {
	var logFile string
	flag.StringVar(&logFile, "logFile", "logs/latest.log", "the log file to watch")
	flag.StringVar(&token, "token", "", "A user's push bullet token")
	flag.StringVar(&userName, "user", "", "A user to watch for")

	flag.Parse()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Println("event:", ev)
				if ev.IsModify() || ev.IsCreate() {
					go readFile(logFile)
				} else if ev.IsRename() {
					handleFileRename(watcher, logFile)
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	addWatcher(watcher, logFile)

	<-done

	watcher.Close()
}

func readFile(fname string) {
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buf := make([]byte, 62)
	stat, err := os.Stat(fname)
	start := stat.Size() - 62
	_, err = file.ReadAt(buf, start)
	if err == nil {
		s := string(buf[:])
		watchMessage := strings.Join([]string{userName, "joined the game\n"}, " ")
		joined := strings.Contains(s, watchMessage)
		if joined {
			go sendPushNotification(watchMessage)
		}
		log.Printf("%s\n joined %v", buf, joined)
	}

}

func handleFileRename(w *fsnotify.Watcher, logFile string) {
	removeWatcher(w, logFile)
	waitForFileReady(logFile)
	addWatcher(w, logFile)
}

func waitForFileReady(fileName string) {

	for {
		time.Sleep(1 * time.Millisecond)
		exists := fileExists(fileName)
		if exists {
			break
		}
	}
}

func fileExists(fileName string) (exists bool) {
	exists = false
	if _, err := os.Stat(fileName); err == nil {
		log.Printf("file exists")
		exists = true
	}
	return
}

func removeWatcher(w *fsnotify.Watcher, logFile string) {
	log.Printf("Removing watcher for %s", logFile)
	w.RemoveWatch(logFile)
}

func addWatcher(w *fsnotify.Watcher, logFile string) {
	err := w.Watch(logFile)
	if err != nil {
		log.Fatal(err)
	}
}

func sendPushNotification(message string) {

	pb := pushbullet.New(token)

	n := requests.NewNote()
	n.Title = "Minecraft"
	n.Body = message

	// Send the note via Pushbullet.
	if _, err := pb.PostPushesNote(n); err != nil {
		log.Printf("error: %s\n", err)
		return
	}
}
