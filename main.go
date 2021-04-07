package main

import (
	"bytes"
	"flag"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/go-rod/rod"
)

func saveAsImage(filename string, b []byte) error {
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return err
	}

	out, _ := os.Create(filename)
	defer out.Close()

	err = jpeg.Encode(out, img, nil)
	if err != nil {
		return err
	}
	return nil

}

func main() {
	var outputFile, url string
	var help bool

	flag.StringVar(&url, "u", "", "The japscan url")
	flag.StringVar(&outputFile,
		"o",
		"japscan_image_"+strconv.FormatInt(time.Now().UnixNano(), 10)+".jpeg",
		"Name of the output file (the downloaded image)")
	flag.BoolVar(&help, "h", false, "Show this help")

	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if url == "" {
		log.SetFlags(0)
		log.Fatal("invalid url, see -h")
	}

	var wg sync.WaitGroup

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd("https://cdn.statically.io/img/c.japscan.se*", func(h *rod.Hijack) {
		h.MustLoadResponse()
		if err := saveAsImage(outputFile, h.Response.Payload().Body); err != nil {
			log.Fatalf("error while saving the image: %v", err)
		}
		wg.Done()
	})

	wg.Add(1)
	go func() {
		router.Run()
	}()

	browser.MustPage(url).MustWaitRequestIdle()
	wg.Wait()
}
