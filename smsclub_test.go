package smsclub

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var (
	testMethods = map[methodAPI]testMethod{
		mBalance: testMethod{
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "1034.17<br/>0")
			}),
			"1034.17 + 0.00",
		},
		mSend: testMethod{
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "=IDS START=<br/>000002<br/>000003<br/>=IDS END=<br/>")
			}),
			"000002 + 000003",
		},
		mStatus: testMethod{
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "=IDS START=<br/>ID_1: STATE<br/>ID_2: STATE<br/>=IDS END=<br/>")
			}),
			"STATE + STATE",
		},
	}
	testSMSer SMSer // It inits in TestNew()
)

type testMethod struct {
	hfnc http.HandlerFunc
	want string
}

func (t testMethod) newServer() *httptest.Server {
	return httptest.NewServer(t.hfnc)
}

func testFindURL(m methodAPI) string {
	return fmt.Sprintf(mapper[m])
}

func testError(t *testing.T, err error) {
	if err != nil {
		t.Errorf(err.Error())
	}
}

func testResult(t *testing.T, got, want string) {
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNew(t *testing.T) {
	testSMSer = New("user", "pass")
	testResult(t, fmt.Sprintf("%s", testSMSer), "user pass 0")
}

func TestBalance(t *testing.T) {
	tm := testMethods[mBalance]
	ts, want := tm.newServer(), tm.want
	defer ts.Close()

	mapper[mBalance] = ts.URL
	makeURL = testFindURL

	bln, cre, err := testSMSer.Balance()
	testError(t, err)

	got := strings.Join([]string{fmt.Sprintf("%.2f", bln), fmt.Sprintf("%.2f", cre)}, " + ")
	testResult(t, got, want)
}

func TestLifeTime(t *testing.T) {
	err := testSMSer.LifeTime(5 * time.Minute)
	testError(t, err)
	testResult(t, fmt.Sprintf("%s", testSMSer), "user pass 5")
	_ = testSMSer.LifeTime(0 * time.Minute)
	testResult(t, fmt.Sprintf("%s", testSMSer), "user pass 0")
}

func TestSend(t *testing.T) {
	tm := testMethods[mSend]
	ts, want := tm.newServer(), tm.want
	defer ts.Close()

	mapper[mSend] = ts.URL
	makeURL = testFindURL

	res, err := testSMSer.Send("Test", "Test", "0123456789")
	testError(t, err)

	got := strings.Join(res, " + ")
	testResult(t, got, want)
}

func TestStatus(t *testing.T) {
	tm := testMethods[mStatus]
	ts, want := tm.newServer(), tm.want
	defer ts.Close()

	mapper[mStatus] = ts.URL
	makeURL = testFindURL

	res, err := testSMSer.Status()
	testError(t, err)

	got := strings.Join(res, " + ")
	testResult(t, got, want)
}
