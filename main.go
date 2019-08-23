package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/proxy"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func initializeProxySwitcher(c *colly.Collector) {
	// Rotate two socks5 proxies

	proxy1 := "socks5://95.110.230.142:50337" //USA 83ms
	proxy2 := "socks5://178.197.249.213:1080" //USA 99ms
	//proxy3 := "socks5://104.238.97.230" //USA 54ms

	rp, err := proxy.RoundRobinProxySwitcher(proxy1, proxy2)
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)
}

func initializeLimits(c *colly.Collector) {
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*fddb.info*",
		Parallelism: 3, // Approx: 45 minute Runtime
		RandomDelay: time.Second * 1,
	})
}

func main() {

	log.Printf("Started")

	mongo := getClient()
	err := mongo.Ping(context.Background(), readpref.Primary())
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected!")
	}

	c0 := colly.NewCollector()

	//initializeProxySwitcher(c0)
	initializeLimits(c0)

	productGroups := extractProductGroups(c0)

	for _, group := range productGroups {
		fmt.Printf("%v\n", group)
	}

	log.Printf("Try fetching Products")
	c1 := c0.Clone() //Same settings but new events
	//products := extractProductsFromGroup(c1, productGroups[0])

	fmt.Printf("Products Amount: %v", calculateAmountOfProducts(c1, productGroups))

	//counter := 0

	/*
		for _, product := range products {
			counter++
			c2 := c0.Clone()
			info := extractProductInformation(c2, product)
			id := insertFoodProduct(mongo, info)
			log.Printf("Number: %v, Inserted: %v", counter, id)
		}*/

	//c2 := c0.Clone()
	//product := extractProductInformation(c2, products[11])
	//product := extractProductInformation(c2, ProductLink{Name: "Milk...", Link: "https://fddb.info/db/de/lebensmittel/balade_so_light/index.html"})
	//fmt.Printf("%# v", pretty.Formatter(product))

}

func calculateAmountOfProducts(c *colly.Collector, productGroups []ProductGroup) int {
	amount := 0
	for _, group := range productGroups {
		c1 := c.Clone()
		amount += len(extractProductsFromGroup(c1, group))
	}
	return amount
}
