package main

import "fmt"
import "sync"
import "github.com/pion/ion-sfu/pkg/dcplug"
import log "github.com/pion/ion-log"
import "github.com/pion/webrtc/v3"
import "time"
import "encoding/json"

var lock sync.RWMutex
var count int = 0
var terminate bool = false
var subscribers map[string]dcplug.SendMessageFunc

type CounterMessage  struct {
	Count int
}

func backgroundTask() {
	ticker := time.NewTicker(1 * time.Second)
	for _ = range ticker.C {
		data,err := json.Marshal( CounterMessage { Count : count } )
		if err != nil {
			panic(err)
		}
		msg := webrtc.DataChannelMessage {
			IsString : true,
			Data : data,
		}
		lock.Lock()
		for _,sf := range subscribers{
			sf( msg )
		}
		lock.Unlock()
		count = count + 1
		if terminate {
			return
		}
	}
}

func Launch() {
	lock.Lock()
	defer lock.Unlock()
	terminate = false
	subscribers = make(map[string]dcplug.SendMessageFunc)
	log.Infof("Launching Plugin")
	go backgroundTask()
}

func Terminate() {
	lock.Lock()
	defer lock.Unlock()
	terminate = true
	log.Infof("Terminating Plugin")
}

func StartSession( sid string ) {
	fmt.Printf("Starting Session in plugin %v", sid)
	lock.Lock()
	defer lock.Unlock()
	log.Infof("Plugin %v session started", sid)
}

func EndSession( sid string) {
	lock.Lock()
	defer lock.Unlock()
	log.Infof("Ending Session in plugin %v", sid)
}

func AddSubscriber( sid string, uid string, msg_hndlr dcplug.SendMessageFunc) {
	lock.Lock()
	defer lock.Unlock()
	subscribers[sid + uid] = msg_hndlr
}

func RemoveSubscriber( sid string, uid string) {
	lock.Lock()
	defer lock.Unlock()
	delete(subscribers, sid + uid)
}

func ProcessMessage( sid string, uid string, msg webrtc.DataChannelMessage ) {
	log.Infof("Got message from %v %v", sid, uid);
	if msg.IsString {
	   log.Infof("message %v", string(msg.Data));
	}
}
