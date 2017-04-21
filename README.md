# smsclub
Golang (>= go1.7) package for sending SMS via **SMS Club**. See [Connection to the SMS-gateway, API](https://smsclub.mobi/en/api) for details.

## It's simple!

Install package:
```
$ go get -u github.com/runningmaster/smsclub
```

Send SMS to friends:
```
sms, _ := smsclub.New(
	smsclub.User("user"),
	smsclub.Token("user_token"),
	smsclub.Sender("alpha_name"),
)
sms.Send("Hello dudes!", "380673408275", "380975243263")
```
