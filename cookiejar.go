package openwechat

import (
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
	"unsafe"
)

// Jar is a struct which as same as cookiejar.Jar
// cookiejar.Jar's fields are private, so we can't use it directly
type Jar struct {
	PsList cookiejar.PublicSuffixList

	// mu locks the remaining fields.
	mu sync.Mutex

	// Entries is a set of entries, keyed by their eTLD+1 and subkeyed by
	// their name/Domain/path.
	Entries map[string]map[string]entry

	// nextSeqNum is the next sequence number assigned to a new cookie
	// created SetCookies.
	NextSeqNum uint64
}

// AsCookieJar unsafe convert to http.CookieJar
func (j *Jar) AsCookieJar() http.CookieJar {
	return (*cookiejar.Jar)(unsafe.Pointer(j))
}

func fromCookieJar(jar http.CookieJar) *Jar {
	return (*Jar)(unsafe.Pointer(jar.(*cookiejar.Jar)))
}

func NewJar() *Jar {
	jar, _ := cookiejar.New(nil)
	return fromCookieJar(jar)
}

type entry struct {
	Name       string
	Value      string
	Domain     string
	Path       string
	SameSite   string
	Secure     bool
	HttpOnly   bool
	Persistent bool
	HostOnly   bool
	Expires    time.Time
	Creation   time.Time
	LastAccess time.Time

	// seqNum is a sequence number so that Jar returns cookies in a
	// deterministic order, even for cookies that have equal Path length and
	// equal Creation time. This simplifies testing.
	seqNum uint64
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

func getWebWxDataTicket(cookies []*http.Cookie) (string, error) {
	cookieGroup := CookieGroup(cookies)
	cookie, exist := cookieGroup.GetByName("webwx_data_ticket")
	if !exist {
		return "", ErrWebWxDataTicketNotFound
	}
	return cookie.Value, nil
}
