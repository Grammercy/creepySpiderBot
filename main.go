package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

var badUrls = 0

func main() {
  go http.ListenAndServe("localhost:6060", nil)
  ch := make(chan colly.HTMLElement, 20)
  ch2 := make(chan colly.HTMLElement, 20)
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
    colly.MaxDepth(-1),
    colly.UserAgent("pls dont block pookie this is a school project"),
  )


  go func () {
    fmt.Println("running handler")
    go urlHandlerAndScheduler(ch)
    go siteWritingHandler(ch2, w)
    fmt.Println("ending handler")
    flag = true
  }()

  c.OnError(func(r *colly.Response, err error) {
		if r.StatusCode == http.StatusTooManyRequests {
			fmt.Println("Received 429 slow down?")
			time.Sleep(10 * time.Second) 
			r.Request.Retry()
			return
		}
    fmt.Println(err)
	})

  c.OnHTML("html", func(e *colly.HTMLElement) {
    // fmt.Println(e)
    ch2 <- *e
  })

  c.OnHTML("a[href]", func(e *colly.HTMLElement) {
    if e.Request == nil {
      return
    }
    ch <- *e
	})

  // c.Visit("https://6d6ldeoyebpfbivviclcauaejm.srv.us/")
  // c.Visit("https://en.wikipedia.org/")
  c.Visit("https://old.reddit.com")
  // c.Visit("")
  c.Wait()
  time.Sleep(1000 * time.Millisecond)
  close(ch)

  
  for !flag {
    time.Sleep(10 * time.Millisecond)
  }
  w.Flush()

  
}

func urlHandlerAndScheduler(chIn chan colly.HTMLElement) {
  for {
    select {
    case e := <- chIn:
      link := e.Request.AbsoluteURL(e.Attr("href"))
      filePath, err := cannonizeUrlAndEnsureDirExists(link)
      if err != nil {
        continue
      }
      // fmt.Println("2")
      _, err = os.Stat(filePath)
      if !os.IsNotExist(err) {
        // fmt.Println("continuing on " + filename)
        continue
      }
      time.Sleep(10 * time.Millisecond)
      go e.Request.Visit(e.Attr("href"))
    }
  }
}

func siteWritingHandler(chIn chan colly.HTMLElement, w *bufio.Writer) {
  for {
    select {
    case e := <- chIn:
      link := e.Request.AbsoluteURL(e.Request.URL.String())
      w.Write([]byte(link + "\n"))
      filePath, err := cannonizeUrlAndEnsureDirExists(link)
      if err != nil {
        panic(err)
      }
      filePath += ".html"
      err = os.WriteFile(filePath, []byte(e.Response.Body), 0644)
      fmt.Println("written to ", filePath)
      if err != nil {
        panic(err)
      }
    }
  }
}

func cannonizeUrlAndEnsureDirExists(link string) (string, error) {
  filename := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(link, "/.", "/"), "//", "/"), "/", "_")
  if link == "" || link == "javascript:void(0)" || link == "javascript: void 0;" {
    // fmt.Println(thing.Req)
    return "", errors.New("javascript url found")
  }
  // fmt.Println("1")
  tld, secondLevelDomain, subdomain, path := urlToPath(link)
  dir := "./sites/" + filepath.Join(".", tld, secondLevelDomain)
  if subdomain != "" {
    dir = filepath.Join(dir, subdomain)
  }
  err := os.MkdirAll(dir, os.ModePerm)
  if err != nil {
    panic(err)
  }
  if path != "" {
    filename = strings.ReplaceAll(path, "/", "_")
  }
  // fmt.Println("3")
  filePath := filepath.Join(dir, filename)
  if !strings.Contains(filePath, "sites") {
    fmt.Println(filePath)
  }
  return filePath, nil

} 

func urlToPath(startingUrl string) (string, string, string, string) {
  parsedUrl, _ := url.Parse(startingUrl)
  linkParts := strings.Split(parsedUrl.Hostname(), ".")
  if len(linkParts) < 2 {
    fmt.Println("Link is shitty " +  startingUrl)
    return "bad", "bad", "bad", string(badUrls)
  }
  tld := linkParts[len(linkParts)-1]
	secondLevelDomain := linkParts[len(linkParts)-2]
	subdomain := ""
	if len(linkParts) > 2 {
		subdomain = strings.Join(linkParts[:len(linkParts)-2], ".")
	}
  return tld, secondLevelDomain, subdomain, parsedUrl.Path
}
