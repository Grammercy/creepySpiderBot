package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gocolly/colly/v2"
)

func main() {
  fmt.Println("Hello")
  linkCounts := make(map[string]int)
  ch := make(chan string, 100)
  running := true
  flag := false
  f, err := os.Create("visitedSites.txt")
	if err != nil {
		fmt.Println("Error creating file", err)
		panic(err)
	}
	defer f.Close()
  w := bufio.NewWriter(f)

  c := colly.NewCollector(
    colly.MaxDepth(3),
    colly.AllowedDomains("en.wikipedia.org"),
    colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
  )

  go func () {
    fmt.Println("running handler")
    for running || len(ch) > 0 {
      link := <-ch 
      linkCounts[link]++
    }
    fmt.Println("ending handler")
    flag = true
  }()

  c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
    ch <- link
    
		go e.Request.Visit(link)
	})

  c.Visit("https://en.wikipedia.org")
  c.Wait()
  time.Sleep(1000 * time.Millisecond)
  close(ch)

  running = false
  
  linkRankings := make(map[int]string) 
  for !flag {
    time.Sleep(10 * time.Millisecond)
  }
  fmt.Println("starting conversion")
  for link, pointingToLink := range linkCounts {
    linkRankings[pointingToLink] += link + ", "
  }
  fmt.Println("converted to linkCounts")
  
  keys := make([]int, 0, len(linkRankings))
  for k := range linkRankings {
    keys = append(keys, k)
  }
  
  sort.Sort(sort.Reverse(sort.IntSlice(keys)))
  fmt.Println("sorted")
  result := make([][]string, len(keys))
  for _, k := range keys {
    arr := make([]string, 2)
    arr[0] = strconv.Itoa(k)
    arr[1] = linkRankings[k]
    result = append(result, arr)
  }
  fmt.Println("converted")
  str := ""
  for _, e := range result {
    str += fmt.Sprint(e) +"\n"
  }
  
  w.WriteString(str)
  w.Flush()

  
  fmt.Println(len(linkCounts))
}
