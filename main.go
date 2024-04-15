package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

const (
	maxHops    = 6
	hitlerLink = "https://en.wikipedia.org/wiki/Adolf_Hitler"
)

type linkQueue struct {
	links []string
}

func (q *linkQueue) enqueue(link string) {
	q.links = append(q.links, link)
}

func (q *linkQueue) dequeue() string {
	if len(q.links) == 0 {
		return ""
	}
	link := q.links[0]
	q.links = q.links[1:]
	return link
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the Wikipedia link to start searching from: ")
	startLink, _ := reader.ReadString('\n')
	startLink = strings.TrimSpace(startLink)

	path := searchHitler(startLink)
	if path != nil {
		fmt.Println("Path to Hitler page:", strings.Join(path, " -> "))
	} else {
		fmt.Println("Hitler not found within", maxHops, "hops")
	}
}

func searchHitler(startLink string) []string {
	queue := linkQueue{}
	visited := make(map[string]bool)
	paths := make(map[string][]string)
	hops := make(map[string]int) 
	queue.enqueue(startLink)
	visited[startLink] = true
	paths[startLink] = []string{startLink}
	hops[startLink] = 1

	for len(queue.links) > 0 {
		currentLink := queue.dequeue()
		if hops[currentLink] > maxHops {
			return nil
		}
		if skipLink(currentLink) {
			continue
		}
		//fmt.Println("Checking:", currentLink)
		resp, err := http.Get(currentLink)
		if err != nil {
			fmt.Println("Error fetching page:", err)
			continue
		}
		defer resp.Body.Close()

		tokenizer := html.NewTokenizer(resp.Body)
		for {
			tokenType := tokenizer.Next()
			if tokenType == html.ErrorToken {
				break
			}

			if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken {
				token := tokenizer.Token()
				if token.Data == "a" {
					for _, attr := range token.Attr {
						if attr.Key == "href" && strings.HasPrefix(attr.Val, "/wiki/") {
							newLink := "https://en.wikipedia.org" + attr.Val
							if newLink == hitlerLink {
								fmt.Println("Found Hitler at", hitlerLink)
								paths[newLink] = append(paths[currentLink], hitlerLink)
								return paths[hitlerLink]
							}
							if !visited[newLink] {
								visited[newLink] = true
								paths[newLink] = append(paths[currentLink], newLink)
								queue.enqueue(newLink)
								hops[newLink] = hops[currentLink] + 1 // Increment the hop count for the new link
							}
						}
					}
				}
			}
		}
	}
	return nil
}

func skipLink(link string) bool {
	// Skip certain pages
	return (link == "https://en.wikipedia.org/wiki/Main_Page") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Wikipedia:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/File:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Template:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Template_talk:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Portal:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Special:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Talk:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Help:") ||
		strings.HasPrefix(link, "https://en.wikipedia.org/wiki/Category:")
}
