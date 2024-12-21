package openwechat

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// Jar is a struct which as same as cookiejar.Jar
// cookiejar.Jar's fields are private, so we can't use it directly
type Jar struct {
	jar   *cookiejar.Jar
	hosts map[string]*url.URL
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (j *Jar) UnmarshalJSON(bytes []byte) error {
	var cookies map[string][]*http.Cookie
	if err := json.Unmarshal(bytes, &cookies); err != nil {
		return err
	}
	if j.jar == nil {
		j.jar, _ = cookiejar.New(nil)
	}
	for u, cs := range cookies {
		u, err := url.Parse(u)
		if err != nil {
			return err
		}
		j.jar.SetCookies(u, cs)
	}
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (j *Jar) MarshalJSON() ([]byte, error) {
	var cookies = make(map[string][]*http.Cookie)
	for path, u := range j.hosts {
		cookies[path] = append(cookies[path], j.jar.Cookies(u)...)
	}
	return json.Marshal(cookies)
}

func (j *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	if j.hosts == nil {
		j.hosts = make(map[string]*url.URL)
	}
	path := u.Scheme + "://" + u.Host
	if _, exists := j.hosts[path]; !exists {
		j.hosts[path] = u
	}
	j.jar.SetCookies(u, cookies)
}

func (j *Jar) Cookies(u *url.URL) []*http.Cookie {
	return j.jar.Cookies(u)
}

func NewJar() *Jar {
	jar, _ := cookiejar.New(nil)
	return &Jar{
		jar:   jar,
		hosts: make(map[string]*url.URL),
	}
}

// CookieGroup is a group of cookies
type CookieGroup []*http.Cookie

func (c CookieGroup) GetByName(cookieName string) (cookie *http.Cookie, exist bool) {
	for _, cookie := range c {
		if cookie.Name == cookieName {
			return cookie, true
		}
	}
	return nil, false
}

func wxDataTicket(cookies []*http.Cookie) (string, error) {
	cookieGroup := CookieGroup(cookies)
	cookie, exist := cookieGroup.GetByName("webwx_data_ticket")
	if !exist {
		return "", ErrWebWxDataTicketNotFound
	}
	return cookie.Value, nil
}
