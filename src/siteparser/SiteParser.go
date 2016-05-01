package SiteParser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

type Jar struct {
	lk      sync.Mutex
	cookies map[string][]*http.Cookie
}

//type ChildTree struct {
//	Parent *Tree
//	Element []*interface{}
//}

type Tree struct {
	Parent  *Tree
	Child   []*Tree
	Element string
}

var (

	//	Client *http.Client
	Cook http.Cookie
	jar  *Jar
)

func Find(text []byte, template []byte, n int) int {
	m := 0
	len_text := len(text)

	for i := 0; i < len_text; i++ {
		isFound := false
		for j := 0; j < len(template); j++ {
			if i+j+1 > len_text {
				break
			}
			if text[i+j] != template[j] {
				break
			}
			if j == len(template)-1 {
				m++
				if m == n {
					isFound = true
				}
				break
			}
		}
		if isFound {
			return i
		}
	}
	return -1
}

func FindTegBlock(text []byte, classname []byte) []byte {
	class_pos := 0
	for {
		//search keyword "class"
		class_pos = Find(text[class_pos:], []byte("class"), 1) + class_pos
		if class_pos == -1 {
			return []byte("")
		}
		//search the first " as start class definition
		//	fmt.Println(class_pos,"=", string(text[class_pos:class_pos+20]))
		start_class_value := Find(text[class_pos:], []byte("\""), 1)
		if start_class_value == -1 {
			return []byte("")
		}
		start_class_value = start_class_value + class_pos + 1
		//search next " as end class definition
		end_class_value := Find(text[start_class_value+2:], []byte("\""), 1)
		if end_class_value == -1 {
			return []byte("")
		}
		end_class_value = end_class_value + start_class_value + 2
		//		fmt.Println(start_class_value,",",end_class_value,"=",string(text[start_class_value:end_class_value]),"|",string(classname))

		//search classname from start to end points
		if Find(text[start_class_value:end_class_value], classname, 1) != -1 {
			for i := start_class_value; i >= 0; i-- {
				if (string(text[i]) == "<") && (string(text[i+1]) != "/") {

					//search teg name
					end_teg_pos := Find(text[i+1:], []byte(">"), 1)
					end_teg_pos2 := Find(text[i+1:], []byte(" "), 1)
					if end_teg_pos2 < end_teg_pos {
						end_teg_pos = end_teg_pos2
					}
					if end_teg_pos == -1 {
						return []byte("")
					}
					end_teg_pos = end_teg_pos + i + 1
					teg := text[i+1 : end_teg_pos]
					//					fmt.Println("teg=",string(teg))
					close_teg := "/" + string(teg) + ">"
					len_close_teg := len(close_teg)
					//					fmt.Println("close teg=",close_teg)
					//search close teg
					open_teg := 1
					pos := end_teg_pos

					for j := 1; j < 1000; j++ {

						offset := Find(text[pos:], teg, 1)
						if offset == -1 {
							fmt.Println("Error in html page, teg=", string(teg), "do not have close teg")
							return []byte("")
						}
						pos = pos + offset
						//fmt.Println()
						if string(text[pos-1:pos-1+len_close_teg]) == close_teg {
							open_teg--
						} else {
							if string(text[pos-1]) == "<" {
								open_teg++
							}
						}
						//						fmt.Println(open_teg)
						if open_teg == 0 {
							//fmt.Println("teg=",string(text[i:pos+len(teg)+1]))
							return text[i : pos+len(teg)+1]
						}
						pos = pos + len(teg)

					}
					fmt.Println("find more 1000 teg=", string(teg))

					return []byte("")
				}
			}
		}
		class_pos = end_class_value
	}

}

func FindTegBlockByParam(text []byte, param []byte, value []byte, end ...*int) []byte {
	param_pos := 0

	if end != nil {
		param_pos = *end[0]

		if *end[0] >= len(text)-1 {
			//		*end[0]=-1
			return []byte("")
		}
	}
	//	*end[0] = 0
	for {
		//search keyword from param
		param_pos = Find(text[param_pos:], param, 1) + param_pos
		if param_pos == -1 {
			return []byte("")
		}
		//search the first " as start param definition
		//	fmt.Println(class_pos,"=", string(text[class_pos:class_pos+20]))
		start_param_value := Find(text[param_pos:], []byte("\""), 1)
		if start_param_value == -1 {
			return []byte("")
		}
		start_param_value = start_param_value + param_pos + 1
		//search next " as end class definition
		end_param_value := Find(text[start_param_value+2:], []byte("\""), 1)
		if end_param_value == -1 {
			return []byte("")
		}
		end_param_value = end_param_value + start_param_value + 2
		//		fmt.Println(start_class_value,",",end_class_value,"=",string(text[start_class_value:end_class_value]),"|",string(value))

		//search value from start to end points
		if Find(text[start_param_value:end_param_value], value, 1) != -1 {
			for i := start_param_value; i >= 0; i-- {
				if (string(text[i]) == "<") && (string(text[i+1]) != "/") {

					//search teg name
					end_teg_pos := Find(text[i+1:], []byte(">"), 1)
					end_teg_pos2 := Find(text[i+1:], []byte(" "), 1)
					if end_teg_pos2 < end_teg_pos {
						end_teg_pos = end_teg_pos2
					}
					if end_teg_pos == -1 {
						return []byte("")
					}
					end_teg_pos = end_teg_pos + i + 1
					teg := text[i+1 : end_teg_pos]
					//					fmt.Println("teg=",string(teg))
					close_teg := "/" + string(teg) + ">"
					len_close_teg := len(close_teg)
					//					fmt.Println("close teg=",close_teg)
					//search close teg
					open_teg := 1
					pos := end_teg_pos

					for j := 1; j < 1000; j++ {

						offset := Find(text[pos:], teg, 1)
						if offset == -1 {
							fmt.Println("Error in html page, teg=", string(teg), "do not have close teg")
							return []byte("")
						}
						pos = pos + offset
						//fmt.Println()
						if string(text[pos-1:pos-1+len_close_teg]) == close_teg {
							open_teg--
						} else {
							if string(text[pos-1]) == "<" {
								open_teg++
							}
						}
						//						fmt.Println(open_teg)
						if open_teg == 0 {
							//fmt.Println("teg=",string(text[i:pos+len(teg)+1]))
							if end != nil {
								*end[0] = pos + len(teg) + 1
							}
							return text[i : pos+len(teg)+1]
						}
						pos = pos + len(teg)

					}
					fmt.Println("find more 1000 teg=", string(teg))

					return []byte("")
				}
			}
		}
		param_pos = end_param_value
	}

}

func FindTegBlocksByParam(text []byte, param []byte, value []byte) [][]byte {
	var result [][]byte
	var end int = 0
	next := true
	for next {
		block := FindTegBlockByParam(text, param, value, &end)
		if string(block) != "" {
			result = append(result, block)
		} else {
			//			fmt.Println(end,"-",len(text))
			//			fmt.Println(string(text[end:]))
			next = false
		}
	}
	return result
}

func ToDigital(text []byte) string {
	convert := func(r rune) rune {
		switch {
		case r >= '0' && r <= '9':
			return r
		case r == ',':
			return '.'
		}

		return rune(0)
	}
	str := strings.Map(convert, string(text))
	return strings.Replace(str, string(0), "", -1)

}

func CutBefore(text []byte, template []byte, n int) []byte {
	m := 0
	for i := 0; i < len(text); i++ {
		isFound := false
		for j := 0; j < len(template); j++ {
			if text[i+j] != template[j] {
				break
			}
			if j == len(template)-1 {
				m++
				if m == n {
					isFound = true
				}
				break
			}
		}
		if isFound {
			return text[i+len(template):]
		}
	}
	return []byte("")
}

func CutAfter(text []byte, template []byte, n int) []byte {
	m := 0
	for i := 0; i < len(text); i++ {
		isFound := false
		for j := 0; j < len(template); j++ {
			if text[i+j] != template[j] {
				break
			}
			if j == len(template)-1 {
				m++
				if m == n {
					isFound = true
				}
				break
			}
		}
		if isFound {
			return text[:i]
		}
	}
	return []byte("")
}

func GetURL(text []byte, domain string) string {
	text = CutBefore(text, []byte("href=\""), 1)
	text = CutAfter(text, []byte("\""), 1)
	if string(CutBefore([]byte(text), []byte("http"), 1)) == "" {
		text = []byte(domain + string(text))
	}
	return string(text)
}

func NewJar() *Jar {
	jar = new(Jar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

// SetCookies handles the receipt of the cookies in a reply for the
// given URL.  It may or may not choose to save the cookies, depending
// on the jar's policy and implementation.
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lk.Lock()
	jar.cookies[u.Host] = cookies
	jar.lk.Unlock()
}

// Cookies returns the cookies to send in a request for the given URL.
// It is up to the implementation to honor the standard cookie use
// restrictions such as in RFC 6265.
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}

func GetPage(Client *http.Client, url string, blacklist ...[]string) []byte {
	if len(blacklist) > 0 {
		for i := 0; i < len(blacklist[0]); i++ {
			if url == blacklist[0][i] {
				//			fmt.Println("Blacklist url:", url)
				return []byte("")
			}
		}
	}
	//	fmt.Println("Open url:", url)
	response, err := Client.Get(url)
	if err != nil {
		fmt.Println("Error openning url:", url)
	} else {
		defer response.Body.Close()
		//	Cook:=response.Cookies()
		//	fmt.Println(Cook)
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		return contents
	}

	return []byte("")
}

func GetBlocks(text []byte, start []byte, end []byte) [][]byte {
	var result [][]byte
	i := 0
	for {
		block := CutBefore(text, start, i+1)

		block = CutAfter(block, end, 1)
		if string(block) == "" {
			break
		}
		result = append(result, block)
		i++
	}

	return result

}
