package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tebeka/selenium"
)

type Product struct {
	URL, Image, Name, Price string
}

const (
	port = 9515
)

func main() {
	url := os.Getenv("URL")
	var products []Product
	opts := []selenium.ServiceOption{}
	caps := selenium.Capabilities{
		"browserName": "chrome",
		"chromeOptions": map[string]interface{}{
			"args": []string{
				"--headless",
				"--disable-gpu",
				"--no-sandbox",
			},
		},
	}

	service, err := selenium.NewChromeDriverService("/opt/homebrew/bin/chromedriver", port, opts...)
	if err != nil {
		log.Fatalf("Failed to start the Chrome driver server: %s\n", err)
	}
	defer service.Stop()

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Fatalf("Error connecting to WebDriver: %v", err)
	}
	defer wd.Quit()

	wd.Get(url)
	if err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}

	time.Sleep(5 * time.Second)

	productCards, err := wd.FindElements(selenium.ByCSSSelector, ".product-item")
	if err != nil {
		log.Fatalf("Failed to find product cards: %v", err)
	}

	for _, card := range productCards {
		nameEl, _ := card.FindElement(selenium.ByCSSSelector, ".item-heading")
		priceEl, _ := card.FindElement(selenium.ByCSSSelector, ".price")
		linkEl, _ := card.FindElement(selenium.ByCSSSelector, "a")
		imgEl, _ := card.FindElement(selenium.ByCSSSelector, "img")

		name, _ := nameEl.Text()
		price, _ := priceEl.Text()
		url, _ := linkEl.GetAttribute("href")
		img, _ := imgEl.GetAttribute("src")

		products = append(products, Product{
			URL:   url,
			Image: img,
			Name:  name,
			Price: price,
		})
	}
	file, err := os.Create("products.csv")
	if err != nil {
		log.Fatalf("Failed to create output CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"URL", "Image", "Name", "Price"})
	for _, p := range products {
		writer.Write([]string{p.URL, p.Image, p.Name, p.Price})
	}
	fmt.Println("Scraping completed")
}
