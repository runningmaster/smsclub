package smsclub

import (
	"bytes"
	"context"
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

const (
	endPoint   = "https://gate.smsclub.mobi/token/"
	urlSend    = endPoint
	urlStatus  = endPoint + "state.php"
	urlBalance = endPoint + "getbalance.php"
)

var mapURL = map[methodAPI]string{
	mSend:    urlSend,
	mStatus:  urlStatus,
	mBalance: urlBalance,
}

// SMSCluber is interface for https://smsclub.mobi/en/pages/show/api
type SMSCluber interface {
	// Balance returns values for balance and credit.
	Balance() (float64, float64, error)

	// Send sends SMS text message to recipients.
	Send(text string, to ...string) ([]string, error)

	// Status gets list of SMS identifiers and returns statuses for ones.
	Status(ids ...string) ([]string, error)
}

// New returns SMSCluber interface. Minimal *must* options are User(), Token().
func New(options ...func(*option) error) (SMSCluber, error) {
	c := &client{
		&option{},
	}

	var err error
	for i := range options {
		err = options[i](c.option)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

type option struct {
	user     string
	token    string
	sender   string // alphaname
	lifetime time.Duration
	timeout  time.Duration
}

type client struct {
	*option
}

func (c *client) String() string {
	return fmt.Sprintf("%s %s %s %d", c.user, c.token, c.sender, c.lifetime)
}

func (c *client) makeForm(f url.Values) url.Values {
	if f == nil {
		f = url.Values{}
	}

	f.Set("username", c.user)
	f.Set("token", c.token)
	if c.lifetime > 0 {
		f.Set("lifetime", strconv.Itoa(int(c.lifetime.Minutes())))
	}

	return f
}

func (c *client) withRequest(m methodAPI, v url.Values) (*http.Request, error) {
	f := c.makeForm(v)
	r, err := http.NewRequest("POST", mapURL[m], strings.NewReader(f.Encode()))
	if err != nil {
		return nil, err
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return r, nil
}

func (c *client) parseResponse(r io.Reader) ([]string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, s := range strings.Split(string(b), "<br/>") {
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

func (c *client) callAPI(r *http.Request) ([]string, error) {
	if c.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		defer cancel()
		r = r.WithContext(ctx)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil || res == nil {
		return nil, err
	}
	if res != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("smsclub: server returns %d %s", res.StatusCode, res.Status)
	}

	return c.parseResponse(res.Body)
}

// Balance returns values for balance and credit.
func (c *client) Balance() (float64, float64, error) {
	req, err := c.withRequest(mBalance, nil)
	if err != nil {
		return 0.0, 0.0, err
	}

	res, err := c.callAPI(req)
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

// Send sends SMS text message to recipients.
func (c *client) Send(text string, to ...string) ([]string, error) {
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
		"from": []string{c.sender},
		"text": []string{toBase64(toWin1251(text))},
		"to":   []string{strings.Join(to, ";")},
	}

	req, err := c.withRequest(mSend, form)
	if err != nil {
		return nil, err
	}
	return c.callAPI(req)
}

// Status gets list of SMS identifiers and returns statuses for ones.
func (c *client) Status(ids ...string) ([]string, error) {
	form := url.Values{
		"smscid": []string{strings.Join(ids, ";")},
	}

	req, err := c.withRequest(mStatus, form)
	if err != nil {
		return nil, err
	}

	res, err := c.callAPI(req)
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

// User is login of user’s account
func User(name string) func(*option) error {
	return func(o *option) error {
		if name == "" {
			return fmt.Errorf("smsclub: username is empty")
		}
		o.user = name
		return nil
	}
}

// Token is token of user’s account (you can find it in profile).
func Token(val string) func(*option) error {
	return func(o *option) error {
		if val == "" {
			return fmt.Errorf("smsclub: token is empty")
		}
		o.token = val
		return nil
	}
}

// Sender is Sender ID, from which mail-out is perfomed (11 English letters, numbers, spaces).
// See https://my.smsclub.mobi/en/alphanames/index.
func Sender(val string) func(*option) error {
	return func(o *option) error {
		if val == "" {
			return fmt.Errorf("smsclub: sender (alphaName) is empty")
		}
		o.sender = val
		return nil
	}
}

// LifeTime sets life time of SMS, which is specified in minutes.
func LifeTime(d time.Duration) func(*option) error {
	return func(o *option) error {
		if d < 0 {
			return fmt.Errorf("smsclub: invalid duration value %d", d)
		}
		o.lifetime = d
		return nil
	}
}

// Timeout sets timeout for calls Balance, Send and Status.
func Timeout(d time.Duration) func(*option) error {
	return func(o *option) error {
		if d < 0 {
			return fmt.Errorf("smsclub: invalid duration value %d", d)
		}
		o.timeout = d
		return nil
	}
}
