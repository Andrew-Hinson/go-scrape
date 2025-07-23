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
	port            = 9515
	waitTime        = 5 * time.Second
	outputFile      = "products.csv"
	productSelector = "ul[data-elid='product-grid'] > li article"
)

func main() {
	url := getRequiredEnv("URL")

	driver := setupWebDriver()
	defer driver.Quit()

	products := scrapeProducts(driver, url)
	saveToCSV(products, outputFile)

	fmt.Printf("Scraping completed. Found %d products.\n", len(products))
}

func getRequiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Environment variable %s is required", key)
	}
	return value
}

func setupWebDriver() selenium.WebDriver {
	caps := selenium.Capabilities{
		"browserName": "chrome",
		"chromeOptions": map[string]interface{}{
			"args": []string{"--headless", "--disable-gpu", "--no-sandbox"},
		},
	}

	service, err := selenium.NewChromeDriverService("/opt/homebrew/bin/chromedriver", port)
	if err != nil {
		log.Fatalf("Failed to start Chrome driver: %v", err)
	}

	driver, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		service.Stop()
		log.Fatalf("Failed to connect to WebDriver: %v", err)
	}

	return driver
}

func scrapeProducts(driver selenium.WebDriver, url string) []Product {
	if err := driver.Get(url); err != nil {
		log.Fatalf("Failed to load page: %v", err)
	}

	time.Sleep(waitTime)

	elements, err := driver.FindElements(selenium.ByCSSSelector, productSelector)
	if err != nil {
		log.Fatalf("Failed to find products: %v", err)
	}

	products := make([]Product, 0, len(elements))
	for _, element := range elements {
		if product := extractProduct(element); product != nil {
			products = append(products, *product)
		}
	}

	return products
}

func extractProduct(element selenium.WebElement) *Product {
	getText := func(selector string) string {
		if el, err := element.FindElement(selenium.ByCSSSelector, selector); err == nil {
			if text, err := el.Text(); err == nil {
				return text
			}
		}
		return ""
	}

	getAttr := func(selector, attr string) string {
		if el, err := element.FindElement(selenium.ByCSSSelector, selector); err == nil {
			if value, err := el.GetAttribute(attr); err == nil {
				return value
			}
		}
		return ""
	}

	name := getText("h2")
	if name == "" {
		return nil // Skip products without names
	}

	return &Product{
		Name:  name,
		Price: getText("span"),
		URL:   getAttr("a", "href"),
		Image: getAttr("img", "src"),
	}
}

func saveToCSV(products []Product, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	writer.Write([]string{"Name", "Price", "URL", "Image"})

	// Write data
	for _, p := range products {
		writer.Write([]string{p.Name, p.Price, p.URL, p.Image})
	}
}
