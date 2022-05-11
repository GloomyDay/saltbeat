package beater

import (
	"fmt"
	"net"
	"time"

	// "strconv"

	"io"
	"strings"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/logp"

	"github.com/GloomyDay/saltbeat/config"
	"github.com/vmihailenco/msgpack"
)

var socketError error

type Saltbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
	messages         chan map[string]interface{}
	socketConnection *net.UnixConn
}


func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Saltbeat{
		done:   make(chan struct{}),
		config: c,
		messages: make(chan map[string]interface{}),
	}
	return bt, nil
}

// Crazy way to detect if connections if REALY closed ( looks like  calling bt.socketConnection.Close() doesn't mean that connections is realy closed)
func Read(c *net.UnixConn, buffer []byte) bool {
    _ , err := c.Read(buffer)
    if err != nil {
		c.Close()
		logp.Debug("message", "Error in read data from UnixConn, connections is closed.")
        return false
    }
	logp.Debug("message", "Connection to UnixConn is still opened.")
    return true
}

func (bt *Saltbeat) socketReconnect() {
	var err error
	var retryCounter int
	var connectAlive bool
	bt.socketConnection.Close()

	for {
		buffer := make([]byte, 2048)
		connectAlive = Read(bt.socketConnection,buffer)
		if connectAlive == false {
			break
		}
	}
	for {
		logp.Debug("message", "Sleeping for 2 seconds before DialUnix")
		time.Sleep(2 * time.Second)
		bt.socketConnection, err = net.DialUnix("unix", nil,&net.UnixAddr{Name: bt.config.MasterEventPub, Net: "unix"})
		buffer := make([]byte, 2048)
		connectAlive = Read(bt.socketConnection,buffer)
		if connectAlive {
			logp.Debug("message", "I can read from socket after reconnect")
		} else {
			logp.Debug("message", "Can't read from socket after reconnect")
		}
		if err != nil {
			logp.Err(fmt.Sprintf("Error: %s ,reconnecting to socket, %d attempts left,sleeping for 1 second before next attempt", err.Error(), 5 - retryCounter))
			retryCounter += 1
			if retryCounter > 6 {
				logp.Err("The maximum number of reconnect attempts has been reached")
				logp.WTF(err.Error())
			}
			time.Sleep(1 * time.Second)
		} else {
			logp.Debug("message", "Reconnecting is OK")
			break	
		}
	}
}

func (bt *Saltbeat) Run(b *beat.Beat) error {
	var err error

	go func ()  (error){
		logp.Info(fmt.Sprintf("Connecting to socket %s", bt.config.MasterEventPub))
		var err error
		var fullMessage map[string]interface{}

		logp.Info("Opening socket %s", bt.config.MasterEventPub)
		bt.socketConnection, err =  net.DialUnix("unix", nil,&net.UnixAddr{Name: bt.config.MasterEventPub, Net: "unix"})
		if err != nil {
			return err
		}
		for {
			logp.Debug("message", "Waiting for message")
			dec := msgpack.NewDecoder(bt.socketConnection) 
			err = dec.Decode(&fullMessage)
			if err != nil && err == io.EOF{
				bt.socketConnection.Close()
				logp.Err(fmt.Sprintf("Error: %s ,reconnecting to socket", err.Error()))
				bt.socketReconnect()
			} else {
				logp.Debug("message", "Message read")
				bt.messages <- fullMessage	
			}
		}
	}()
	
	logp.Info("saltbeat is running! Hit CTRL-C to stop it.")

	bt.client, err = b.Publisher.Connect()

	if err != nil {
		return err
	}

	for {
		select {
			case <-bt.done:
				return nil
			case message := <-bt.messages:
				var data map[string]interface{}
				skipMessage := false
				resultList := strings.SplitN(string(message["body"].([]uint8)), "\n\n", 2)
				tag := resultList[0]
				for _, s := range bt.config.TagBlackList {
					if strings.Contains(tag, s){
						skipMessage = true
						break
					}  
				}
				if skipMessage{
					logp.Debug("message", fmt.Sprintf("Found blacklisted message with tag: %s ", tag))
					continue
				}
				byteResult := []byte(resultList[1])
				_ = msgpack.Unmarshal(byteResult, &data)
				// Clear the return so we don't show passwords
				data["return"] = ""
				// Drop public keys ( garbage )
				data["pub"] = ""
				logp.Debug("message", fmt.Sprintf("Decoded message: \nTag : %s \ndata: %s\n", tag, data))
				
				logp.Debug("publish", "Publishing event")
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"type":    b.Info.Name,
						"tag":        tag,
						"data":    data,
					},
				}
				bt.client.Publish(event)
		}
	}
}

func (bt Saltbeat) Cleanup(b *beat.Beat) error {
	logp.Info("Closing socket %s", bt.config.MasterEventPub)
	bt.socketConnection.Close()
	return nil
}

func (bt Saltbeat) Stop() {
	close(bt.done)
	close(bt.messages)
}
