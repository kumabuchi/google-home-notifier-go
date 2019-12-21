package notifier

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/barnybug/go-cast"
	"github.com/barnybug/go-cast/controllers"
)

// Notifier is a google-home-notifier-go client
type Notifier struct {
	client *cast.Client
	ctx    context.Context
}

// NewClient makes a connection and create a client
func NewClient(ctx context.Context, host string, port int) (*Notifier, error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	client := cast.NewClient(ips[0], port)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	log.Println("Connected")
	n := &Notifier{client: client, ctx: ctx}
	return n, nil
}

// Set volume
func (n *Notifier) Volume(vol string) error {
	receiver := n.client.Receiver()
	level, _ := strconv.ParseFloat(vol, 64)
	muted := false
	volume := controllers.Volume{Level: &level, Muted: &muted}
	_, err := receiver.SetVolume(n.ctx, &volume)
	return err
}

// Wait for play to become ready
func (n *Notifier) Wait(timeout int) error {
	media, err := n.client.Media(n.ctx)
	if err != nil {
		return err
	}
	wait := 0
	for {
		response, err := media.GetStatus(n.ctx)
		if err != nil {
			return err
		}
		time.Sleep(1 * time.Second)
		if len(response.Status) < 1 {
			break
		}
		wait += 1
		if wait > timeout {
			break
		}
	}
	return nil
}

// Notify sends a message to google home
func (n *Notifier) Notify(text string, language string, waitTimeout int) error {
	baseURL := "https://translate.google.com/translate_tts?ie=UTF-8&q=%s&tl=%s&client=tw-ob"
	u := fmt.Sprintf(baseURL, url.QueryEscape(text), url.QueryEscape(language))
	return n.Play(u, waitTimeout)
}

//Play sound via URL
func (n *Notifier) Play(url string, waitTimeout int) error {
	media, err := n.client.Media(n.ctx)
	if err != nil {
		return err
	}
	n.Wait(waitTimeout)
	contentType := "audio/mpeg"
	item := controllers.MediaItem{
		ContentId:   url,
		StreamType:  "BUFFERED",
		ContentType: contentType,
	}
	_, err = media.LoadMedia(n.ctx, item, 0, true, map[string]interface{}{})
	n.Wait(waitTimeout)
	return err
}

//Stop sound
func (n *Notifier) Stop() error {
	if !n.client.IsPlaying(n.ctx) {
		return nil
	}
	media, err := n.client.Media(n.ctx)
	if err != nil {
		return err
	}
	_, err = media.Stop(n.ctx)
	return err
}

// Quit
func (n *Notifier) Quit() error {
	receiver := n.client.Receiver()
	_, err := receiver.QuitApp(n.ctx)
	return err
}

// Close connection
func (n *Notifier) Close() {
	n.client.Close()
}
