# smsclub
Golang (>= go1.7) package for sending SMS via **SMS Club of Ukraine**. See [Integration, API](https://smsclub.mobi/en/pages/show/api) for details.

## It's simple!

Install package:
```
$ go get -u github.com/runningmaster/smsclub
```

Send SMS to friends:
```
sms := smsclub.New(
	smsclub.User("user"),
	smsclub.Token("user_token"),
	smsclub.Sender("alpha_name"),
)
sms.Send("Hello dudes!", "380673408275", "380975243263")
```