package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"./siteparser"
)

var (
	Client    *http.Client
	Domain    string
	err       error
	tv        TV
	ChannelID map[string]string
)

type TV struct {
	XMLName      xml.Name  `xml:"tv"`
	Generator    string    `xml:"generator-info-name,attr"`
	Created      string    `xml:"created-by,attr"`
	ChannelList  []Channel `xml:"channel"`
	ProgrammList []Program `xml:"programme"`
}

type Channel struct {
	XMLName xml.Name          `xml:"channel"`
	Id      string            `xml:"id,attr"`
	DN      DisplayNameStruct `xml:"display-name"`
	Url     string            `xml:"url"`
}

type DisplayNameStruct struct {
	XMLName     xml.Name `xml:"display-name"`
	DisplayName string   `xml:",chardata"`
	Lang        string   `xml:"lang,attr"`
}

type Program struct {
	XMLName xml.Name    `xml:"programme"`
	Channel string      `xml:"channel,attr"`
	Start   string      `xml:"start,attr"`
	Stop    string      `xml:"stop,attr"`
	TS      TitleStruct `xml:"title"`
	DS      DescStruct  `xml:"desc"`
}

type TitleStruct struct {
	XMLName xml.Name `xml:"title"`
	Title   string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr"`
}

type DescStruct struct {
	XMLName xml.Name `xml:"desc"`
	Desc    string   `xml:",chardata"`
	Lang    string   `xml:"lang,attr"`
}

func main() {
	fmt.Println(time.Now(), "Starting...")
	XmlFilePtr := flag.String("xml-file", "./spored_tv.xml", "output file")
	flag.Parse()
	firstPage := "http://www.spored.tv/"
	Domain = "http://www.spored.tv"
	//create http client with cooks
	jar := SiteParser.NewJar()
	Client = &http.Client{Jar: jar}

	//get channelId map
	input, err := ioutil.ReadFile("channel-id.json")
	if err != nil {
		fmt.Printf("error read json file: %v\n", err)
	}
	err = json.Unmarshal(input, &ChannelID)
	if err != nil {
		fmt.Printf("error parsing json file: %v\n", err)
	}
	//fmt.Println("Channel ID:", ChannelID)

	//get first page
	page := SiteParser.GetPage(Client, firstPage)

	//get list of chanel from left block of site
	tv.ChannelList = GetChannelList(page)
	tv.Generator = "Alternet"
	tv.Created = firstPage
	for currentChanell := 0; currentChanell < len(tv.ChannelList); currentChanell++ {
		//get channel page
		page := SiteParser.GetPage(Client, tv.ChannelList[currentChanell].Url)
		channel := tv.ChannelList[currentChanell].Id
		//get urls for days
		daysURL := GetDaysURL(page)
		for currentDay := 0; currentDay < len(daysURL); currentDay++ {
			var pr Program

			page := SiteParser.GetPage(Client, daysURL[currentDay])
			day := GetDaySelected(page)

			//get container with list of programs
			block := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("ScheduleItemsContainer"))

			//get list of programs
			items := SiteParser.FindTegBlocksByParam(block, []byte("class"), []byte("ScheduleItem"))
			for currentItem := 0; currentItem < len(items); currentItem++ {
				pr.DS.Desc = ""
				pr.TS.Title = ""
				pr.TS.Lang = "sl"
				pr.DS.Lang = "sl"
				pr.Channel = channel

				//parsing program title
				pr.TS.Title = GetTitle(items[currentItem])
				if pr.TS.Title == "" {
					fmt.Println("Error parsing title:", string(items[currentItem]))
				}
				//parsing program description
				pr.DS.Desc = GetDescription(items[currentItem])
				if pr.DS.Desc == "" {
					fmt.Println("Error parsing description:", string(items[currentItem]))
				}

				//parsing start time
				time := GetStartTime(items[currentItem])
				if time == "" {
					fmt.Println("Error parsing time:", string(items[currentItem]))
				} else {
					pr.Start = day + time + "00 +0200"
				}

				//parsing stop time
				if currentItem < len(items)-1 {
					time := GetStartTime(items[currentItem+1])
					if time == "" {
						fmt.Println("Error parsing stop time :", string(items[currentItem+1]))
					} else {
						pr.Stop = day + time + "00 +0200"
					}
				}
				tv.ProgrammList = append(tv.ProgrammList, pr)
				//	fmt.Println(pr.TS.Title, "\t\t", pr.Start, "\t", pr.Stop, "\t\t", pr.DS.Desc)
			}
		}
	}

	//	generate XML file
	output, err := xml.MarshalIndent(tv, " ", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	//add xml header
	output = []byte(xml.Header + string(output))

	//	write XML to file
	err = ioutil.WriteFile(*XmlFilePtr, output, 0644)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

}

func GetStationHeader(page []byte) []byte {
	block := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("StationHeader"))
	stationHeader := SiteParser.GetBlocks(block, []byte("<h1>"), []byte("</h1>"))
	return stationHeader[0]
}

func GetChannelList(page []byte) []Channel {
	block := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("MenuContainer"))
	items := SiteParser.GetBlocks(block, []byte("<a"), []byte("/a>"))
	var channelList []Channel
	for i := 0; i < len(items); i++ {
		var new_channel Channel
		new_channel.DN.DisplayName = string(SiteParser.GetBlocks(items[i], []byte("title=\""), []byte("\""))[0])

		//id from json file if exist
		fmt.Println("Name:", new_channel.DN.DisplayName, "Value: ", ChannelID[new_channel.DN.DisplayName])
		if ChannelID != nil {
			if ChannelID[new_channel.DN.DisplayName] != "" {
				new_channel.Id = ChannelID[new_channel.DN.DisplayName]
			} else {
				new_channel.Id = new_channel.DN.DisplayName
			}
		}

		new_channel.Url = SiteParser.GetURL(items[i], Domain)
		new_channel.DN.Lang = "sl"
		channelList = append(channelList, new_channel)
		//		fmt.Println(string(cl[i].url))
	}
	return channelList
}

func GetDaysURL(page []byte) []string {
	block := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("StationDays"))
	items := SiteParser.GetBlocks(block, []byte("<a"), []byte("/a>"))
	var urls []string
	for i := 0; i < len(items); i++ {
		urls = append(urls, SiteParser.GetURL(items[i], Domain))
	}
	return urls
}

func GetDaySelected(page []byte) string {
	block := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("DaySelected"))
	urlWithDate := SiteParser.GetURL(block, Domain)
	//	fmt.Println(urlWithDate)
	re := regexp.MustCompile(".*([0-9][0-9])-([0-9][0-9])-([0-9][0-9][0-9][0-9])")
	dateArray := re.FindStringSubmatch(urlWithDate)
	//	fmt.Println(dateArray)
	date := ""
	if len(dateArray) == 4 {
		date = dateArray[3] + dateArray[2] + dateArray[1]
	}
	return date

}

func GetTitle(page []byte) string {
	var result = ""
	blockForTitle := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("ProgramDescriptionLink"))
	title := SiteParser.GetBlocks(blockForTitle, []byte(">"), []byte("<"))
	//if title exist
	if len(title) > 0 {
		result = string(title[0])
	} else {

		title = SiteParser.GetBlocks(page, []byte("/span>"), []byte("<"))
		if len(title) > 0 {
			result = strings.TrimSpace(string(title[0]))
		} else {
			fmt.Println("error parsing title:", string(page), "=")
		}
	}
	return result

}

func GetDescription(page []byte) string {
	var result = ""
	blockForDescription := SiteParser.FindTegBlockByParam(page, []byte("id"), []byte("DivProgramDescription_"))
	description := SiteParser.GetBlocks(blockForDescription, []byte(">"), []byte("</div>"))
	//if description exist
	if len(description) > 0 {
		result = string(ClearHTMLTag(description[0]))
		//result = string(description[0])
	}
	return strings.TrimSpace(result)
}

func GetStartTime(page []byte) string {
	var result = ""
	block := SiteParser.GetBlocks(page, []byte("<span"), []byte("</span>"))
	//if description exist
	if len(block) == 1 {
		re := regexp.MustCompile(".*([0-9][0-9]):([0-9][0-9])")
		timeArray := re.FindStringSubmatch(string(block[0]))
		if len(timeArray) == 3 {
			result = timeArray[1] + timeArray[2]
		}
	} else {
		fmt.Println("error parsing start time:", string(page), "=")
	}
	return result
}

//func ClearHTMLTag(text []byte) []byte {
//	//list of tags for deleting
//	tagList := []string{"<i>", "</i>", "br />", "<br>", "&nbsp;"}
//	for i := 0; i < len(tagList); i++ {
//		next := true
//		for next {
//			//find teg
//			n := SiteParser.Find(text, []byte(tagList[i]), 1)
//			//delete teg from text, if exist
//			if n != -1 {
//				if tagList[i] == "&nbsp;" {
//					text = []byte(string(text[:n]) + " " + string(text[n+len(tagList[i]):]))
//				} else {
//					//copy(text[n-1:], text[n+1+len(tagList[i]):])
//					//text = text[:len(text)-len(tagList[i])-1]
//					text = []byte(string(text[:n]) + string(text[n+len(tagList[i]):]))
//				}
//			} else {
//				next = false
//			}
//		}
//	}
//	return text
//}

func ClearHTMLTag(text []byte) []byte {
	//list of tags for deleting
	tagList := []string{"<i>", "</i>", "<br />", "<br>", "&nbsp;", "<a.*a>"}
	for i := 0; i < len(tagList); i++ {
		re := regexp.MustCompile(tagList[i])
		if tagList[i] == "&nbsp;" {
			text = []byte(re.ReplaceAllLiteralString(string(text), " "))
		} else {
			text = []byte(re.ReplaceAllLiteralString(string(text), ""))
		}
	}
	return text
}
