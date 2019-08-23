package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

//Product API
type Product struct {
	Name              string   `json:"name"`
	Description       string   `json:"description"`
	ProductGroup      string   `json:"productGroup"`
	Producer          string   `json:"producer"`
	DataSource        string   `json:"dataSource"`
	CalorificValue    string   `json:"calorificValue"`
	Calories          string   `json:"calories"`
	Protein           string   `json:"protein"`
	Carbohydrates     string   `json:"carbohydrates"`
	ThereofSugar      string   `json:"thereofSugar"`
	Fat               string   `json:"fat"`
	Fiber             string   `json:"fiber"`
	BreadUnits        string   `json:"breadUnits"`
	Measure           string   `json:"measure"`
	PortionSizeNumber string   `json:"PortionSizeNumber"`
	PortionSizeName   string   `json:"PortionSizeName"`
	PortionSizeAmount string   `json:"PortionSizeAmount"`
	Tags              []string `json:"tags"`
}

//CSelection Type
type CSelection struct {
	*goquery.Selection
}

func extractProductInformation(c *colly.Collector, productLink ProductLink) Product {

	targetProduct := Product{}
	targetProduct.Name = productLink.Name
	targetProduct.DataSource = productLink.Link

	selectionChannel := make(chan *goquery.Selection, 1)

	go func() {
		c.OnHTML("div[id]", func(e *colly.HTMLElement) {
			if e.Attr("id") == "content" {
				selectionChannel <- e.DOM
			}
		})
		c.Visit(targetProduct.DataSource)
	}()

	selection := CSelection{<-selectionChannel}

	productGroupChan := make(chan string, 1)
	descriptionChan := make(chan string, 1)
	producerChan := make(chan string, 1)
	tagsChan := make(chan []string, 1)
	calorificValueChan := make(chan string, 1)
	caloriesChan := make(chan string, 1)
	proteinChan := make(chan string, 1)
	carbohydratesChan := make(chan string, 1)
	thereofSugarChan := make(chan string, 1)
	fatChan := make(chan string, 1)
	fiberChan := make(chan string, 1)
	breadUnitsChan := make(chan string, 1)
	measureChan := make(chan string, 1)
	portionSizeNumberChan := make(chan string, 1)
	portionSizeNameChan := make(chan string, 1)
	portionSizeAmountChan := make(chan string, 1)

	//START IN PARALLEL
	go selection.extractLeftHandedData(productGroupChan, descriptionChan, producerChan, tagsChan)
	go selection.extractNutritionalInformation(calorificValueChan, caloriesChan, proteinChan, carbohydratesChan, thereofSugarChan, fatChan, fiberChan, breadUnitsChan)
	go selection.extractPortionInformations(measureChan, portionSizeNumberChan, portionSizeNameChan, portionSizeAmountChan)

	log.Printf("Wait for Finish")

	//WAIT FOR FINISH
	targetProduct.Description = <-descriptionChan
	targetProduct.ProductGroup = <-productGroupChan
	targetProduct.Producer = <-producerChan

	targetProduct.CalorificValue = <-calorificValueChan
	targetProduct.Calories = <-caloriesChan
	targetProduct.Protein = <-proteinChan
	targetProduct.Carbohydrates = <-carbohydratesChan
	targetProduct.ThereofSugar = <-thereofSugarChan
	targetProduct.Fat = <-fatChan
	targetProduct.Fiber = <-fiberChan
	targetProduct.BreadUnits = <-breadUnitsChan
	targetProduct.Measure = <-measureChan
	targetProduct.PortionSizeNumber = <-portionSizeNumberChan
	targetProduct.PortionSizeName = <-portionSizeNameChan
	targetProduct.PortionSizeAmount = <-portionSizeAmountChan
	targetProduct.Tags = <-tagsChan

	return targetProduct
}

func (s *CSelection) extractLeftHandedData(productGroup chan string, description chan string, producer chan string, tags chan []string) {

	go s.extractProductGroup(productGroup)
	go s.extractDescription(description)
	go s.extractProducer(producer)
	go s.extractTags(tags)

}

func (s *CSelection) extractProductGroup(productGroup chan string) {

	sl := s.firstSelectionContains("a", "href", "produktgruppen")
	if sl == nil {
		productGroup <- ""
	}

	productGroup <- sl.Text()
}

func (s *CSelection) extractDescription(description chan string) {

	sl := s.firstSelectionContains("p", "class", "lidesc")
	found := ""
	if sl != nil {
		found = sl.Text()
	}
	if strings.Contains(found, "Noch keine Beschreibung für dieses Produkt.") {
		found = ""
	}

	description <- found
}

func (s *CSelection) extractProducer(producer chan string) {

	found := s.firstSelectionContains("a", "href", "hersteller").Text()

	producer <- found
}

func (s *CSelection) extractTags(tags chan []string) {

	found := []string{}
	ignore := false
	s.Find("td[valign='top'] > span, a").Each(func(i int, sl *goquery.Selection) {
		class, classexits := sl.Attr("class")
		if strings.Contains(sl.Text(), "Angaben noch nicht bestätigt.") || strings.Contains(sl.Text(), "Melde einen Fehler") || strings.Contains(sl.Text(), "Foto hochladen") || strings.Contains(sl.Text(), "korrigiere die") {
			return
		}
		if strings.Contains(sl.Text(), "Schreibe eine Bewertung") {
			ignore = true
		}
		if classexits {
			if class == "lghtlnk" {
				return
			}
		}
		if ignore {
			return
		}
		if len(sl.Text()) == 0 {
			return
		}
		found = append(found, sl.Text())
	})

	tags <- found
}

func (s *CSelection) extractNutritionalInformation(calorificValue chan string, calories chan string, protein chan string, carbohydrates chan string, thereofSugar chan string, fat chan string, fiber chan string, breadUnits chan string) {

	allSidrowPairs := s.extractSidrowNeighbors()

	calorificValue <- allSidrowPairs["Brennwert"]
	calories <- allSidrowPairs["Kalorien"]
	protein <- allSidrowPairs["Protein"]
	carbohydrates <- allSidrowPairs["Kohlenhydrate"]
	thereofSugar <- allSidrowPairs["davon Zucker"]
	fat <- allSidrowPairs["Fett"]
	fiber <- allSidrowPairs["Ballaststoffe"]
	breadUnits <- allSidrowPairs["Broteinheiten"]

}

func (s *CSelection) extractSidrowNeighbors() map[string]string {

	pairs := make(map[string]string)
	s.Find("div[class='sidrow']").Each(func(i int, sl *goquery.Selection) {
		//log.Printf("%v: %v", sl.Text(), sl.Next().Text())
		pairs[sl.Text()] = sl.Next().Text()
	})
	return pairs
}

func (s *CSelection) extractPortionInformations(measure chan string, portionSizeNumber chan string, portionSizeName chan string, portionSizeAmount chan string) {

	rMeasure := ""
	rPortionSizeNumber := ""
	rPortionSizeName := ""
	rPortionSizeAmount := ""

	result := s.firstSelectionContains("a", "class", "servb")
	if result == nil {
		measure <- rMeasure
		portionSizeNumber <- rPortionSizeNumber
		portionSizeName <- rPortionSizeName
		portionSizeAmount <- rPortionSizeAmount
		return
	}

	r, err := regexp.Compile(`\(\d{2,4}\s.*\)`)
	if err != nil {
		panic(err)
	}

	//Extract Informations from String

	portionStr2 := s.Find("a[class='servb']").Last().Text()

	portionStr := getLargerAmountPortionString(result.Text(), portionStr2, r)

	//\(\d{2,4}\s.*\) --> extracts (XXXX xxx)
	//\d{1,6} --> extracts first number
	//\D* --> extracts words with whitespace replace aftwards whitespace with ""

	amountAndMeasure := r.FindString(portionStr)
	nameAndAmount := strings.Replace(portionStr, amountAndMeasure, "", 1)

	//Send back
	measure <- extractLetters(amountAndMeasure)
	portionSizeNumber <- extractNumeric(nameAndAmount)
	portionSizeName <- extractLetters(nameAndAmount)
	portionSizeAmount <- extractNumeric(amountAndMeasure)
}

func getLargerAmountPortionString(portionStr1 string, portionStr2 string, r *regexp.Regexp) string {
	amountAndMeasure1 := r.FindString(portionStr1)
	portionSizeAmount1, _ := strconv.Atoi(extractNumeric(amountAndMeasure1))

	amountAndMeasure2 := r.FindString(portionStr2)
	portionSizeAmount2, _ := strconv.Atoi(extractNumeric(amountAndMeasure2))

	if portionSizeAmount1 > portionSizeAmount2 {
		return portionStr1
	}
	return portionStr2

}

func (s *CSelection) firstSelectionContains(selector string, attr string, substring string) *CSelection {
	found := CSelection{}
	s.Find(selector).Each(func(i int, sl *goquery.Selection) {
		if val, exits := sl.Attr(attr); exits == true {
			if strings.Contains(val, substring) && found == (CSelection{}) {
				found = CSelection{sl}
			}
		}
	})
	if found == (CSelection{}) {
		return nil
	}
	return &found
}

func extractNumeric(strWithNumbers string) string {
	retVal := ""
	for _, char := range strWithNumbers {
		if unicode.IsDigit(char) {
			retVal = retVal + string(char)
		}
	}
	return retVal
}

func extractLetters(strWithNumbers string) string {
	retVal := ""
	for _, char := range strWithNumbers {
		if unicode.IsLetter(char) {
			retVal = retVal + string(char)
		}
	}
	return retVal
}
