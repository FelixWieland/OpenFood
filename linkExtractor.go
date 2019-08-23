package main

import (
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

//ProductGroup API
type ProductGroup struct {
	Name string
	Link string
}

//ProductLink API
type ProductLink struct {
	Name string
	Link string
}

const fddbInfo = "https://fddb.info"
const fddbInfoProductGroups = "https://fddb.info/db/de/produktgruppen/produkt_verzeichnis/index.html"

//extractProductGroups etracts product groups
func extractProductGroups(c *colly.Collector) []ProductGroup {
	//https://fddb.info/db/de/produktgruppen/produkt_verzeichnis/index.html

	groupChannel := make(chan ProductGroup)
	productGroups := []ProductGroup{}

	go func() {
		c.OnHTML("div[class]", func(e *colly.HTMLElement) {
			if e.Attr("class") == "standardcontent" {
				log.Printf("Visiting ProductGroups")
				e.DOM.Find("a").Each(func(i int, s *goquery.Selection) {
					href, _ := s.Attr("href")
					groupChannel <- ProductGroup{Link: fddbInfo + "/" + href, Name: s.Text()}
				})
				close(groupChannel)
			}
		})
		c.Visit(fddbInfoProductGroups)
	}()

	for msg := range groupChannel {
		if len(msg.Name) == 0 {
			continue
		}
		productGroups = append(productGroups, msg)
	}

	return productGroups
}

func extractProductsFromGroup(c *colly.Collector, productGroup ProductGroup) []ProductLink {

	productsChannel := make(chan ProductLink)
	productLinks := []ProductLink{}

	go func() {

		c.OnHTML("div[class]", func(e *colly.HTMLElement) {
			if e.Attr("class") == "leftblock" {
				e.DOM.Find("a").Each(func(i int, s *goquery.Selection) {

					href, _ := s.Attr("href")

					if strings.Contains(href, "fddb.info/db/de/lebensmittel/") {
						productsChannel <- ProductLink{Link: href, Name: s.Text()}
					}
				})
				close(productsChannel)
			}
		})
		c.Visit(productGroup.Link)
	}()

	for msg := range productsChannel {
		productLinks = append(productLinks, msg)
	}

	return productLinks
}
