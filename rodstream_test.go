package rodstream_test

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-rod/rod"
	rodstream "github.com/navicstein/rod-stream"
)

func TestMustPrepareLauncher(t *testing.T) {
	var l = rodstream.MustPrepareLauncher(rodstream.LauncherArgs{
		UserMode: false,
	})

	var extensionId []string

	if value, ok := l.Flags["whitelisted-extension-id"]; ok {
		extensionId = value
	} else if value, ok := l.Flags["allowlisted-extension-id"]; ok {
		extensionId = value
	} else {
		t.Error("Neither whitelisted-extension-id nor allowlisted-extension-id is set")
	}

	if extensionId[0] != rodstream.ExtensionId {
		t.Errorf("Extension is invalid")
	}

}

func TestMustCreatePage(t *testing.T) {
	browser := createBrowser()
	pageInfo := rodstream.MustCreatePage(browser)
	if pageInfo.CapturePage.MustInfo().Title != "Video Streamer" {
		t.Errorf("Page title is invalid, got '%s'", pageInfo.CapturePage.MustInfo().Title)
	}

}

func TestMustGetStream(t *testing.T) {
	url := "https://www.youtube.com/watch?v=Jl8iYAo90pE"
	browser := createBrowser()
	constraints := &rodstream.StreamConstraints{
		Audio:              true,
		Video:              true,
		MimeType:           "video/webm;codecs=vp9,opus",
		AudioBitsPerSecond: 128000,
		VideoBitsPerSecond: 2500000,
		BitsPerSecond:      8000000,
		FrameSize:          1000,
	}

	page := browser.MustPage()
	page.MustNavigate(url).MustWaitRequestIdle()

	page.MustElement(".ytp-large-play-button").MustClick()

	pageInfo := rodstream.MustCreatePage(browser)
	streamCh := make(chan string, 1024)

	if err := rodstream.MustGetStream(pageInfo, *constraints, streamCh); err != nil {
		log.Panicln(err)
	}

	time.AfterFunc(time.Minute, func() {
		if err := rodstream.MustStopStream(pageInfo); err != nil {
			log.Panicln(err)
		}
		browser.MustClose()
		os.Exit(0)
	})

	fpath := "/tmp/video-test.webm"
	videoFile, err := os.Create(fpath)
	if err != nil {
		panic(err)
	}

	for {
		b64Str, ok := <-streamCh
		if !ok {
			close(streamCh)
			break
		}

		b := rodstream.Parseb64(b64Str)
		videoFile.Write(b)
	}

	t.Logf("recording stopped, video available here: %s", fpath)
}

func createBrowser() *rod.Browser {
	var l = rodstream.MustPrepareLauncher(rodstream.LauncherArgs{
		UserMode: false,
	}).
		Bin("/usr/bin/brave-browser").
		MustLaunch()

	browser := rod.New().ControlURL(l).
		NoDefaultDevice().
		MustConnect()

	return browser
}
