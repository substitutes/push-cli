package parser

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	log "github.com/sirupsen/logrus"
	"github.com/substitutes/substitutes/lookup"
	"github.com/substitutes/substitutes/parser"
	"github.com/substitutes/substitutes/structs"
	"io"
	"strings"
)

func processEncoding(data []byte) io.Reader {
	body := make([]byte, len(data))
	iconv.Convert(data, body, "iso-8859-1", "utf-8")
	return bytes.NewReader(body)
}

// GetClasses grabs a "Druck_Kla.htm" file and extracts the name of the classes
func GetClasses(data []byte) []string {
	doc, err := goquery.NewDocumentFromReader(processEncoding(data))
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	var classes []string
	doc.Find("td a").Each(func(i int, sel *goquery.Selection) {
		val, exists := sel.Attr("href")
		if exists {
			classes = append(classes, val)
		}
	})
	return classes
}

// GetSubstitutes parses a given class file
func GetSubstitutes(data []byte) structs.SubstituteResponse {
	doc, err := goquery.NewDocumentFromReader(processEncoding(data))
	if err != nil {
		log.Fatal("Failed to open file: ", err)
	}
	var substitutes []structs.Substitute
	doc.Find("tbody").Last().Find("tr").Each(func(i int, sel *goquery.Selection) {
		if i != 0 {
			var v structs.Substitute
			sel.Find("td font").Each(func(i int, sel *goquery.Selection) {
				t := strings.Replace(sel.Text(), "\n", "", -1)
				switch i {
				case 0:
					v.Classes = sel.Find("b").Text()
					break
				case 1:
					v.Hour = t
					break
				case 2:
					l := lookup.New()
					raw := strings.Split(t, "?")[0]
					v.Teacher = l.Get(raw)
					v.TeacherInitials = raw
					break
				case 3:
					v.Subject = t
					break
				case 4:
					v.Room = strings.Replace(t, "?", " => ", 1)
					break
				case 5:
					v.Type = t
					break
				case 6:
					v.Notes += t
					break
				}
			})
			substitutes = append(substitutes, v)

		}
	})

	parsedDate, err := parser.ParseUntisTime(doc.Find("body > center > font > table:nth-child(1) > tbody > tr:nth-child(2) > td:nth-child(3)").First().Text())
	if err != nil {
		log.Fatal("Failed to parse date: ", err)
	}

	parsedUpdated, err := parser.ParseUntisTime(doc.Find("table").First().Find("tr").Last().Find("td").Last().Text())
	if err != nil {
		log.Fatal("Failed to parse date: ", err)
	}

	meta := structs.SubstituteMeta{
		Date:    parsedDate,
		Class:   strings.Replace(doc.Find("center font font font").First().Text(), "\n", "", -1),
		Updated: parsedUpdated,
	}
	return structs.SubstituteResponse{Meta: meta, Substitutes: substitutes}
}
