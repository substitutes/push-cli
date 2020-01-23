package parser

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
	log "github.com/sirupsen/logrus"
	"github.com/substitutes/substitutes/parser"
	"github.com/substitutes/substitutes/structs"
	"io"
	"regexp"
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
	var extended bool
	var substitutes []structs.Substitute
	doc.Find("table").Last().Remove()
	doc.Find("table").Last().Find("tr").Each(func(i int, sel *goquery.Selection) {
		if i != 0 {
			var v structs.Substitute
			count := len(sel.Find("td").Nodes)
			if count >= 10 /* Not working ,_, */ {
				extended = true
				sel.Find("td").Each(func(i int, sel *goquery.Selection) {
					t := strings.Replace(sel.Text(), "\n", "", -1)
					t = strings.TrimSpace(t)
					switch i {
					// Parse the HTML table into the struct
					case 0:
						var err error
						v.Date, err = parser.ParseUntisTime(t)
						if err != nil {
							break
						}
						break
					case 1:
						v.Hour = t
						break
					case 2:
						v.Day = t
						break
					case 3:
						v.Teacher = t
						v.TeacherInitials = t
						break
					case 4:
						v.Time = t
						break
					case 5:
						v.Subject = t
						break
					case 6:
						v.Type = t
						if t == "Vertretung" {
							v.Type = "Substitute"
						}
						break
					case 7:
						v.Notes = t
						break
					case 8:
						v.Classes = t
						break
					case 9:
						v.Room = strings.Replace(t, "?", " => ", 1)
						break
					case 10:
						v.After = t
						break
					case 11:
						// Check if there is content
						v.Cancelled = len(strings.Replace(t, " ", "", -1)) != 0
						break
					case 12:
						matched, err := regexp.MatchString("x|X", t)
						if err != nil {
							log.Fatal("Failed to compile parser RegEx")
						}
						v.New = matched
						break
					case 13:
						v.Reason = t
						break
					case 14:
						v.Counter = t
						break
					}
				})
			} else { // Alternative parser, deprecated
				extended = false
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
						v.Teacher = strings.Replace(t, "?", " => ", 1)
						v.TeacherInitials = strings.Replace(t, "?", " => ", 1)
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

			}
			substitutes = append(substitutes, v)

		}
	})

	parsedDate, err := parser.ParseUntisDate(doc.Find("center font font b").First().Text())
	if err != nil {
		log.Fatal("Failed to parse date: ", err)
	}

	parsedUpdated, err := parser.ParseUntisTime(doc.Find("table").First().Find("tr").Last().Find("td").Last().Text())
	if err != nil {
		log.Fatal("Failed to parse date: ", err)
	}

	meta := structs.SubstituteMeta{
		Extended: extended,
		Date:     parsedDate,
		Class:    strings.Replace(doc.Find("center font font font").First().Text(), "\n", "", -1),
		Updated:  parsedUpdated,
	}
	return structs.SubstituteResponse{Meta: meta, Substitutes: substitutes}
}
