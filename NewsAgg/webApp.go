package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

var wg sync.WaitGroup

type Sitemapindex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Loc        []string `xml:"url>loc"`
	Changefreq []string `xml:"url>changefreq"`
}

type NewsMap struct {
	Loc        string
	Changefreq string
}

type NewsAggPage struct {
	Title string
	News  map[int]NewsMap
}

func newsRoutine(c chan News, Location string) {
	defer wg.Done()
	var n News
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	c <- n
}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s Sitemapindex

	resp, _ := http.Get("https://www.washingtonpost.com/arcio/sitemap/master/index/")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	newsMap := make(map[int]NewsMap)
	queue := make(chan News, 30)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go newsRoutine(queue, s.Locations[i])
	}
	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx, v := range elem.Loc {
			newsMap[idx] = NewsMap{v, elem.Changefreq[idx]}
		}
	}
	p := NewsAggPage{Title: "Todays News", News: newsMap}
	t, _ := template.ParseFiles("webApp.html")
	fmt.Println(t.Execute(w, p))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello User, Ready to see some news!!")
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/newsIndex", newsAggHandler)
	http.ListenAndServe(":9000", nil)
}
