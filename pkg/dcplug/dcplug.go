package dcplug

import (
        "fmt"
	"plugin"
	"sync"
	log "github.com/pion/ion-log"
	"github.com/pion/webrtc/v3"
)


// DCPluginHost represents a Datachannel Plugin as loaded on the SFU.
type DCPluginHost struct {
   sync.RWMutex
   plug              *plugin.Plugin
   launch             plugin.Symbol
   start_session     plugin.Symbol
   add_subscriber    plugin.Symbol
   process_message   plugin.Symbol
   terminate         plugin.Symbol
   end_session       plugin.Symbol
   remove_subscriber plugin.Symbol
}

type SendMessageFunc func(msg webrtc.DataChannelMessage) error

// NewDCPluginHost creates a new plugin on the host
// side. It loads and resolves all the symbols
func NewDCPluginHost(path string) (*DCPluginHost, error) {
	plug, err := plugin.Open(path)
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	launch, err := plug.Lookup("Launch")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	start_session, err := plug.Lookup("StartSession")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	add_subscriber, err := plug.Lookup("AddSubscriber")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	process_message, err := plug.Lookup("ProcessMessage")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	terminate, err := plug.Lookup("Terminate")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	end_session, err := plug.Lookup("EndSession")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	remove_subscriber, err := plug.Lookup("RemoveSubscriber")
	if err != nil {
           return nil, fmt.Errorf("Load DC Plugin %v: %v", path, err)
	}
	return &DCPluginHost{
		plug:                plug,
		launch:              launch,
		start_session:       start_session,
		add_subscriber:      add_subscriber,
		process_message:     process_message,
		terminate:           terminate,
		end_session:         end_session,
		remove_subscriber:   remove_subscriber,
	}, nil
}

func (plug *DCPluginHost) Launch() {
	log.Infof("Launching Plugin")
	plug.Lock()
	defer plug.Unlock()
	plug.launch.(func())()
	log.Infof("Plugin Launched")
}

func (plug *DCPluginHost) Terminate() {
	plug.Lock()
	defer plug.Unlock()
	plug.terminate.(func())()
}

func (plug *DCPluginHost) StartSession( sid string ) {
	log.Infof("Starting Session")
	plug.Lock()
	defer plug.Unlock()
	plug.start_session.(func(sid string))( sid )
	log.Infof("Session Started")
}

func (plug *DCPluginHost) EndSession( sid string ) {
	plug.Lock()
	defer plug.Unlock()
	plug.end_session.(func(sid string))( sid )
}

func (plug *DCPluginHost) AddSubscriber( sid string, uid string, message_handler SendMessageFunc  ) {
	plug.Lock()
	defer plug.Unlock()
	plug.add_subscriber.(func(sid string, uid string, msg_hndlr SendMessageFunc))( sid, uid, message_handler )
}

func (plug *DCPluginHost) RemoveSubscriber( sid string, uid string ) {
	plug.Lock()
	defer plug.Unlock()
	plug.remove_subscriber.(func(sid string, uid string))( sid, uid )
}

func (plug *DCPluginHost) ProcessMessage( sid string, uid string, msg webrtc.DataChannelMessage ) {
	plug.Lock()
	defer plug.Unlock()
	plug.process_message.(func(sid string, uid string, msg webrtc.DataChannelMessage ))( sid, uid, msg)
}
