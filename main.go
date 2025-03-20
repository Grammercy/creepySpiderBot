package main

import (
	"bufio"
	"fmt"
	"os"
  "path/filepath"
	"strings"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"runtime"
	"time"

	"github.com/gocolly/colly/v2"
)

type Stuff struct {
  Link string;
  Req *colly.Request;
}

func getMemoryStats() {
	for {
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		// Display memory usage information
		fmt.Printf("Alloc = %v MB\tTotalAlloc = %v MB\tSys = %v MB\tNumGC = %v\n", 
			memStats.Alloc / 1024 / 1024,
			memStats.TotalAlloc / 1024 / 1024,
			memStats.Sys / 1024 / 1024,
			memStats.NumGC)
		time.Sleep(2 * time.Second)
	}
}

func main() {
  go http.ListenAndServe("localhost:6060", nil)
  ch := make(chan Stuff, 20)
  ch2 := make(chan colly.HTMLElement, 20)
  running := true
  flag := false
  f, err := os.Create("visitedSites.txt")
  // f2, err := os.Create("queue.txt")
	if err != nil {
		fmt.Println("Error creating file", err)
		panic(err)
	}
	defer f.Close()
  // defer f2.Close()
  w := bufio.NewWriter(f)
  // w2 := bufio.NewWriter(f2)

  c := colly.NewCollector(
    colly.MaxDepth(3),
    colly.UserAgent("pls dont block pookie this is a school project"),
  )


  go func () {
    fmt.Println("running handler")
    for running || len(ch) > 0 {
      // fmt.Println("looping")
      select {
      case thing := <- ch:
        if thing.Req == nil {
          fmt.Println("bad sign, nil requests")
          time.Sleep(100 * time.Millisecond)
          break
        }
        link := thing.Req.AbsoluteURL(thing.Link)
        if link == "" {
          fmt.Println(thing.Req)
          break
        }
        // fmt.Println("1")
        tld, secondLevelDomain, subdomain, path := urlToPath(link)
        dir := filepath.Join(".", tld, secondLevelDomain)
        if subdomain != "" {
			    dir = "./sites/" + filepath.Join(dir, subdomain)
		    }
        // fmt.Println("2")
        err = os.MkdirAll(dir, os.ModePerm)
		    if err != nil {
          panic(err)
        }
        filename := strings.ReplaceAll(link, "/", "_")
        if path != "" {
          filename = strings.ReplaceAll(path, "/", "_")
        }
        // fmt.Println("3")
        filePath := filepath.Join(dir, filename)
        _, err := os.Stat(filePath)
        if !os.IsNotExist(err) {
          continue
        }
        select {
        case e := <- ch2:
          w.Write([]byte(link + "\n"))
          err = os.WriteFile(filePath, []byte(e.Response.Body), 0644)
          fmt.Println("written to ", filePath)
          if err != nil {
            panic(err)
          }
        }
        // fmt.Println("4")

        time.Sleep(1 * time.Millisecond)
        go thing.Req.Visit(thing.Link)
      }
    }
    fmt.Println("ending handler")
    flag = true
  }()

  c.OnHTML("html", func(e *colly.HTMLElement) {
    // fmt.Println(e)
    ch2 <- *e
  })

  c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
    ch <- Stuff{Link: link,Req: e.Request}
	})

  // c.Visit("https://6d6ldeoyebpfbivviclcauaejm.srv.us/")
  c.Visit("https://en.wikipedia.org/")
  // c.Visit("https://old.reddit.com")
  // c.Visit("")
  c.Wait()
  time.Sleep(1000 * time.Millisecond)
  close(ch)

  running = false
  
  for !flag {
    time.Sleep(10 * time.Millisecond)
  }
  w.Flush()

  
}

func urlToPath(startingUrl string) (string, string, string, string) {
  parsedUrl, _ := url.Parse(startingUrl)
  linkParts := strings.Split(parsedUrl.Hostname(), ".")
  if len(linkParts) < 2 {
    panic("Link is shitty " +  startingUrl)
  }
  tld := linkParts[len(linkParts)-1]
	secondLevelDomain := linkParts[len(linkParts)-2]
	subdomain := ""
	if len(linkParts) > 2 {
		subdomain = strings.Join(linkParts[:len(linkParts)-2], ".")
	}
  return tld, secondLevelDomain, subdomain, parsedUrl.Path
}
