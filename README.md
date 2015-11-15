# smsclub
Golang package for sending SMS via **SMS Club of Ukraine**. See [Integration, API](https://smsclub.mobi/en/pages/show/api) for details.

## It's simple!

Install package:
```
$ go get github.com/runningmaster/smsclub
```

Send SMS to friends:
```
sms := smsclub.New("user", "pass")
sms.Send("message text", "alphaname", "380673408275", "380975243263")
```