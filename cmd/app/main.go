package main

// https://www.imdb.com/search/title/?genres=comedy&start=1101&explore=title_type,genres&ref_=adv_nxt

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
	"time"
	"github.com/PuerkitoBio/goquery"
)

var imdbCategoryKeywords []string = []string {
	"comedy",
	"sci-fi",
	"horror",
	"romance",
	"action",
	"thriller",
	"drama",
	"mystery",
	"crime",
	"animation",
	"adventure",
	"fantasy",
	"comedy,romance",
	"action,comedy",
	"superhero",
	"short",
	"war",
	"biography",
	"crime",
	"family",
	"fantasy",
	"game-show",
	"history",
	"music",
	"musical",
	"western",
	"talk-show",
	"sport",
	"reality-tv",
}

func main() {
	ingestMoviesFromIMDB()
	fmt.Println("COMPLETE!")
}

func generateIMDBURLForKeyword(keyword string, start int) string {
	return fmt.Sprintf("https://www.imdb.com/search/title/?title_type=movie&genres=%s&start=%d&explore=title_type,genres&ref_=adv_nxt", keyword, start)
}

func ingestMoviesFromIMDB() {
	for _, keyword := range imdbCategoryKeywords {
		for start := 1; start < 97192; start += 50 {
			ingestionUrl := generateIMDBURLForKeyword(keyword, start)
			fmt.Printf("requesting url: %s", ingestionUrl)

			var html *string = nil
			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()

			actions := []chromedp.Action {
				chromedp.Navigate(ingestionUrl),
				chromedp.WaitVisible(`div.lister-list`),
				chromedp.Sleep(523 * time.Millisecond),
			}

			// this pre-planned step will get html from DOM
			actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
				node, err := dom.GetDocument().Do(ctx)
				if err != nil {
					fmt.Println(err)
					return err
				}

				data, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
				if err != nil {
					fmt.Println(err)
					return err
				}

				html = &data
				return err
			}))

			if err := chromedp.Run(ctx, actions...); err != nil {
				fmt.Errorf("could not navigate to page: %v", err)
			}

			// process the HTML here...
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(*html))
			if err != nil {
				log.Println(err)
				continue
			}

			if doc != nil {
				// Find the media items
				doc.Find(".lister-list .lister-item .lister-item-content").Each(func(i int, s *goquery.Selection) {
					// For each item found, get the band and title
					title := strings.TrimSpace(s.Find(".lister-item-header a").Text())
					year := strings.TrimSpace(s.Find(".lister-item-header span.lister-item-year").Text())
					filmRating := strings.TrimSpace(s.Find("p.text-muted span.certificate").Text())
					genreList := strings.TrimSpace(s.Find("p.text-muted span.genre").Text())
					duration := strings.TrimSpace(s.Find("p.text-muted span.runtime").Text())
					userRating := strings.TrimSpace(s.Find(".ratings-bar strong").Text())

					fmt.Printf("title: %s year: %s film rating: %s genres: %s duration: %s, user rating: %s\n",
						title,
						year,
						filmRating,
						genreList,
						duration,
						userRating)
				})
			}
		}
	}
}
