package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var intervalFlag = flag.Float64("interval", 1, "Interval between each download (sec)")
var httpClient = &http.Client{}

func main() {
	flag.Parse()
	url := flag.Arg(0)

	m := regexp.MustCompile(`^https?://i9i9.to/c/(\d+)`).FindStringSubmatch(url)
	if m == nil {
		log.Fatalln("Invalid i9i9.to URL")
	}
	id := m[1]

	fmt.Printf("Scraping %s...", url)
	l, err := getNumberOfImages(url)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("done")
	fmt.Printf("Found %d images.\n", l)

	for i := 1; i <= l; {
		imgUrl := fmt.Sprintf(`https://i.i9i9.to/image/%s/%d.jpg`, id, i)
		fmt.Printf("Downloading %s...", imgUrl)
		err := download(imgUrl)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Retry...")
		} else {
			fmt.Println("done")
			i++
		}
		if *intervalFlag > 0 {
			fmt.Printf("Waiting for %f seconds...", *intervalFlag)
			time.Sleep(time.Duration(*intervalFlag) * time.Second)
			fmt.Println("OK.")
		}
	}
}

func getNumberOfImages(url string) (int, error) {
	res, err := httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return 0, err
	}

	l := doc.Find(`div.thumb-container`).Length()
	if l == 0 {
		return 0, fmt.Errorf("Can't find images")
	}

	return l, nil
}

func download(rawurl string) error {
	filename, err := fileNameOf(rawurl)
	if err != nil {
		return err
	}
	resp, err := http.Get(rawurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

var reInPath = regexp.MustCompile("[^/]+$")

func fileNameOf(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", err
	}
	file := reInPath.FindString(url.Path)
	if file == "" {
		return "", fmt.Errorf("Filename not found: %s", rawurl)
	}
	return file, nil
}
