package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	//	"strings"
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
	XmlFilePtr := flag.String("xml-file", "./www_tvsporedi_si.xml", "output file")
	flag.Parse()
	firstPage := "http://www.tvsporedi.si/spored.php"
	Domain = "http://www.tvsporedi.si/spored.php"
	//create http client with cooks
	jar := SiteParser.NewJar()
	Client = &http.Client{Jar: jar}

	//get channelId map
	input, err := ioutil.ReadFile("channel-id.json")
	if err != nil {
		fmt.Printf("error read json file: %v\n", err)
	} else {
		err = json.Unmarshal(input, &ChannelID)
		if err != nil {
			fmt.Println("error parsing json file: %v\n", err)
		}
	}

	//get first page
	page := SiteParser.GetPage(Client, firstPage)

	//get list of chanel from left block of site
	tv.ChannelList = GetChannelList(page)
	tv.Generator = "Alternet"
	tv.Created = firstPage
	for currentChanell := 0; currentChanell < len(tv.ChannelList); currentChanell++ { //len(tv.ChannelList); currentChanell++ {
		//get channel page
		page := SiteParser.GetPage(Client, tv.ChannelList[currentChanell].Url)
		channel := tv.ChannelList[currentChanell].Id
		//	fmt.Println("Channel: ", channel, "\n url: ", tv.ChannelList[currentChanell].Url)

		//get days
		t := time.Now()
		blocks := SiteParser.FindTegBlocksByParam(page, []byte("class"), []byte("tab_content"))
		for currentDay := 0; currentDay < len(blocks); currentDay++ {
			//get list of programs
			var pr Program
			pr.TS.Lang = "sl"
			pr.DS.Lang = "sl"
			pr.Channel = channel

			var day string
			switch currentDay {
			case 0:
				day = t.Format("20060102")
			case 1:
				day = t.Add(time.Duration(24) * time.Hour).Format("20060102")
			case 2:
				day = t.Add(time.Duration(48) * time.Hour).Format("20060102")
			}
			next := true
			for next {
				//get start time program
				blockForTime := SiteParser.FindTegBlockByParam(blocks[currentDay], []byte("class"), []byte("time"))
				if string(blockForTime) == "" {
					next = false
					//			fmt.Println("Exit from day: \n", string(blocks[currentDay]))
					break
				}
				start_time := GetStartTime(blockForTime)
				pr.Start = day + string(start_time) + "00 +0200"

				//cut block with time
				blocks[currentDay] = SiteParser.CutBefore(blocks[currentDay], blockForTime, 1)
				//		fmt.Println("After cut block with time: \n", string(blocks[currentDay]))

				//get end time of program
				blockForTime = SiteParser.FindTegBlockByParam(blocks[currentDay], []byte("class"), []byte("time"))
				if string(blockForTime) == "" {
					pr.Stop = pr.Start
				} else {
					stop_time := GetStartTime(blockForTime)
					pr.Stop = day + string(stop_time) + "00 +0200"
				}
				//get title
				blockForTitle := SiteParser.FindTegBlockByParam(blocks[currentDay], []byte("class"), []byte("prog"))
				pr.TS.Title = GetTitle(blockForTitle)

				//cut block with title
				blocks[currentDay] = SiteParser.CutBefore(blocks[currentDay], blockForTitle, 1)
				//		fmt.Println("After cut block with title: \n", string(blocks[currentDay]))

				//get description
				blockForDescription := SiteParser.CutAfter(blocks[currentDay], []byte("</div>"), 1)
				pr.DS.Desc = GetDescription(blockForDescription)

				//cut block with description
				if string(blockForDescription) != "" {
					blocks[currentDay] = SiteParser.CutBefore(blocks[currentDay], blockForDescription, 1)
					//			fmt.Println("After cut block with description: \n", string(blocks[currentDay]))
				}

				tv.ProgrammList = append(tv.ProgrammList, pr)
				//		fmt.Println(pr.Start, "\n", pr.TS.Title, "\n", pr.DS.Desc)
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

func GetChannelList(page []byte) []Channel {

	items := SiteParser.FindTegBlocksByParam(page, []byte("class"), []byte("squared"))

	var channelList []Channel
	for i := 0; i < len(items); i++ {

		var new_channel Channel
		dn := SiteParser.GetBlocks(items[i], []byte("<span>"), []byte("</span>"))
		if dn != nil {
			new_channel.DN.DisplayName = string(dn[0])
		} else {
			new_channel.DN.DisplayName = ""
			fmt.Println("Error parsing channel name:", string(items[i]))
		}
		//fmt.Println("Name:", new_channel.DN.DisplayName, "Value: ", ChannelID[new_channel.DN.DisplayName])

		//id from json file if exist
		if ChannelID != nil {
			if ChannelID[new_channel.DN.DisplayName] != "" {
				new_channel.Id = ChannelID[new_channel.DN.DisplayName]
			} else {
				new_channel.Id = new_channel.DN.DisplayName
			}
		} else {
			new_channel.Id = new_channel.DN.DisplayName
		}

		new_channel.Url = SiteParser.GetURL(items[i], Domain)
		new_channel.DN.Lang = "sl"
		channelList = append(channelList, new_channel)
		//fmt.Println(string(channelList[i].DN.DisplayName), "\t", string(channelList[i].Url))
	}
	return channelList

}

func GetTitle(page []byte) string {
	var result = ""

	re := regexp.MustCompile(".*<a.*>(.*)</a")
	titleArray := re.FindStringSubmatch(string(page))
	//fmt.Println("title block:", titleArray)
	if len(titleArray) == 2 {
		result = titleArray[1]
	} else {
		re := regexp.MustCompile(".*>(.*)</div")
		titleArray := re.FindStringSubmatch(string(page))
		//	fmt.Println("title block:", titleArray)
		if len(titleArray) == 2 {
			result = titleArray[1]
		} else {
			fmt.Println("error parsing title:", string(page))
		}
	}
	return result

}

func GetDescription(page []byte) string {
	var result = ""

	re := regexp.MustCompile(".*>(.*)")
	descriptionArray := re.FindStringSubmatch(string(page))
	//fmt.Println("description block:", descriptionArray)
	if len(descriptionArray) == 2 {
		result = descriptionArray[1]
	} else {
		//		fmt.Println("error parsing description:", string(page))
	}
	return result

}

func GetStartTime(page []byte) string {
	var result = ""

	re := regexp.MustCompile(".*([0-9][0-9]):([0-9][0-9]).*")
	timeArray := re.FindStringSubmatch(string(page))
	//fmt.Println("time block:", timeArray)
	if len(timeArray) == 3 {
		result = timeArray[1] + timeArray[2]
	} else {
		fmt.Println("error parsing start time:", string(page))
	}
	return result
}

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
