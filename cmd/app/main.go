package main

// https://www.imdb.com/search/title/?genres=comedy&start=1101&explore=title_type,genres&ref_=adv_nxt

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

var imdbCategoryKeywords []string = []string{
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

type Movie struct {
    Title       string   `json:"title"`
    Year        string   `json:"year"`
    FilmRating  string   `json:"filmRating"`
    GenreList   []string `json:"genreList"`
    Duration    string   `json:"duration"`
    UserRating  string   `json:"userRating"`
	Description string   `json:"description"`
}

var Movies map[string]*Movie
var count int = 0

func main() {
	Movies = make(map[string]*Movie)
    ingestMoviesFromIMDB()
    fmt.Println("COMPLETE!")
}

func generateIMDBURLForKeyword(keyword string, start int) string {
    return fmt.Sprintf("https://www.imdb.com/search/Title/?title_type=movie&genres=%s&start=%d&explore=title_type,genres&ref_=adv_nxt", keyword, start)
}

func getCategoryItemCount(keyword string) int {
	ingestionUrl := generateIMDBURLForKeyword(keyword, 1)

	var html *string = nil
	ctx, cancel := chromedp.NewContext(context.Background())

	actions := []chromedp.Action{
		chromedp.Navigate(ingestionUrl),
		chromedp.WaitVisible(`div.lister-list`),
		//chromedp.Sleep(523 * time.Millisecond),
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
	}

	if doc != nil {
		rawTitleCountStr := doc.Find("div.article div.nav div.desc").Find("span").First().Text()
		raw := rawTitleCountStr[8:]
		raw = raw[0:len(raw) - 8]
		raw = strings.ReplaceAll(raw, ",", "")

		val, err := strconv.Atoi(raw)
		if err != nil {
			log.Fatal(err)
		}

		cancel()
		return val
	}

	return -1
}

func ingestMoviesFromIMDB() {
    for _, keyword := range imdbCategoryKeywords {
    	count = getCategoryItemCount(keyword)
    	fmt.Printf("found %d items in category %s\n\n", count, keyword)

        for start := 1; start < count; start += 50 {
			if start > 1 && ((start - 1) % 250 == 0) {
				// status update
				fmt.Printf("\n\n\n\ncategory %s items processed %d out of %d. %.2f%%\n\n\n\n", keyword, start, count, float64(start)/float64(count))
				time.Sleep(time.Duration(1) * time.Second)
			}
			if start > 1 && ((start - 1) % 5000) == 0 {
				// save current data structure to disk
				fmt.Printf("\n\nsaving memory to disk...\n\n")
				exportToJSON()
			}

			ingestMoviePage(keyword, start)
        }
    }

	exportToJSON()
}

func ingestMoviePage(keyword string, start int) {
	ingestionUrl := generateIMDBURLForKeyword(keyword, start)
	fmt.Printf("requesting url: %s", ingestionUrl)

	var html *string = nil
	ctx, cancel := chromedp.NewContext(context.Background())

	actions := []chromedp.Action{
		chromedp.Navigate(ingestionUrl),
		chromedp.WaitVisible(`div.lister-list`),
		//chromedp.Sleep(523 * time.Millisecond),
	}

	// this pre-planned step will get html from DOM
	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		data, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		if err != nil {
			//fmt.Println(err)
			return err
		}

		html = &data
		return nil
	}))

	if err := chromedp.Run(ctx, actions...); err != nil {
		fmt.Errorf("could not navigate to page: %v", err)
	}

	// process the HTML here...
	if html != nil {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(*html))
		if err != nil {
			log.Println(err)
			return
		}

		if doc != nil {
			// Find the media items
			doc.Find(".lister-list .lister-item .lister-item-content").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the band and Title
				title := strings.TrimSpace(s.Find(".lister-item-header a").Text())

				year := strings.TrimSpace(s.Find(".lister-item-header span.lister-item-year").Text())
				year = strings.ReplaceAll(year, "(", "")
				year = strings.ReplaceAll(year, ")", "")
				year = strings.ReplaceAll(year, "I", "")
				year = strings.ReplaceAll(year, "-", "")

				duration := strings.TrimSpace(s.Find("p.text-muted span.runtime").Text())
				userRating := strings.TrimSpace(s.Find(".ratings-bar strong").Text())

				sel := s.Find("p.text-muted")
				filmRating := strings.TrimSpace(sel.Eq(0).Find("span.certificate").Text())
				genreList := strings.Split(strings.TrimSpace(sel.Eq(0).Find("span.genre").Text()), ", ")

				description := strings.TrimSpace(sel.Eq(1).Text())

				movie := &Movie{
					Title:       title,
					Year:        year,
					FilmRating:  filmRating,
					GenreList:   genreList,
					Duration:    duration,
					UserRating:  userRating,
					Description: description,
				}

				key := fmt.Sprintf("%s|%s", title, year)
				if _, ok := Movies[key]; ok {
					fmt.Printf("skipping because already exists in memory. %s\n", key)
					return
				}

				fmt.Println("ADDED KEY:" + key)
				Movies[key] = movie
			})
		}

		cancel()
	} else {
		fmt.Println("no html in page!")
	}
}

func exportToJSON() {
	bytes, err := json.Marshal(Movies)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("writing dataset file with %d bytes...\n", len(bytes))
	if err := ioutil.WriteFile("/output/dataset.json", bytes, 0644); err != nil {
		log.Fatal(err)
	}
}
