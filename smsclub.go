package smsclub

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type methodAPI int

const (
	mSend methodAPI = iota
	mStatus
	mBalance
)

var (
	epoint = "https://gate.smsclub.mobi/http"
	mapper = map[methodAPI]string{
		mSend:    "httpsendsms",
		mStatus:  "httpgetsmsstate",
		mBalance: "httpgetbalance",
	}
	makeURL = func(m methodAPI) string {
		return fmt.Sprintf("%s/%s.php", epoint, mapper[m])
	}
)

// SMSer is interface for https://smsclub.mobi/en/pages/show/api
type SMSer interface {
	// Balance returns values for balance and credit.
	Balance() (float64, float64, error)

	// LifeTime sets life time of SMS, which is specified in minutes.
	LifeTime(d time.Duration) error

	// Send sends SMS text message from "alphaName" to recipients.
	Send(from, text string, to ...string) ([]string, error)

	// Status gets list of SMS identifiers and returns statuses for ones.
	Status(ids ...string) ([]string, error)
}

// New returns SMSer interface
func New(user, pass string) SMSer {
	return &client{
		user: user,
		pass: pass,
	}
}

type client struct {
	user  string
	pass  string
	ltime int
}

func (c *client) makeForm(f url.Values) (url.Values, error) {
	if c.user == "" || c.pass == "" {
		return nil, fmt.Errorf("smsclub: username or password is empty")
	}
	if f == nil {
		f = url.Values{}
	}
	f.Set("username", c.user)
	f.Set("password", c.pass)
	if c.ltime > 0 {
		f.Set("lifetime", strconv.Itoa(c.ltime))
	}
	return f, nil
}

func (c *client) callAPI(m methodAPI, v url.Values) ([]string, error) {
	frm, err := c.makeForm(v)
	if err != nil {
		return nil, err
	}

	res, err := http.PostForm(makeURL(m), frm)
	defer func() {
		err := res.Body.Close()
		if err != nil {
			panic(err)
		}
	}()
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("smsclub: server returns %d %s", res.StatusCode, res.Status)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, s := range strings.Split(string(bts), "<br/>") {
		if strings.Contains(s, "=") || strings.TrimSpace(s) == "" {
			continue
		}
		out = append(out, s)
	}

	if out == nil {
		return nil, io.EOF
	}

	return out, nil
}

func (c *client) Balance() (float64, float64, error) {
	res, err := c.callAPI(mBalance, nil)
	if err != nil {
		return 0.0, 0.0, err
	}

	var bal, cre float64
	if len(res) == 2 {
		bal, _ = strconv.ParseFloat(res[0], 64)
		cre, _ = strconv.ParseFloat(res[1], 64)
	}

	return bal, cre, nil
}

func (c *client) LifeTime(d time.Duration) error {
	if d < 0 {
		return fmt.Errorf("smsclub: invalid duration value %d", d)
	}
	c.ltime = int(d.Minutes())
	return nil
}

func (c *client) Send(from, text string, to ...string) ([]string, error) {
	toBase64 := func(s string) string {
		return base64.StdEncoding.EncodeToString([]byte(s))
	}
	toWin1251 := func(s string) string {
		b := new(bytes.Buffer)
		w := transform.NewWriter(b, charmap.Windows1251.NewEncoder())
		_, _ = w.Write([]byte(s))
		return b.String()
	}

	form := url.Values{
		"from": []string{from},
		"text": []string{toBase64(toWin1251(text))},
		"to":   []string{strings.Join(to, ";")},
	}

	return c.callAPI(mSend, form)
}

func (c *client) Status(ids ...string) ([]string, error) {
	form := url.Values{
		"smscid": []string{strings.Join(ids, ";")},
	}

	res, err := c.callAPI(mStatus, form)
	if err != nil {
		return nil, err
	}

	for i := range res {
		spl := strings.Split(res[i], ":")
		if len(spl) == 2 {
			res[i] = strings.TrimSpace(spl[1])
		}
	}

	return res, nil
}

func (c *client) String() string {
	return fmt.Sprintf("%s %s %d", c.user, c.pass, c.ltime)
}
