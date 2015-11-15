# smsclub
smsclub is a golang package for sending SMS via **SMS Club of Ukraine**.

See [Integration, API](https://smsclub.mobi/en/pages/show/api) for details.

## It's simple!
```
sms := smsclub.New("user", "pass")
ids, err := sms.Send("message text", "alphaname", "380673408275", "380975243263")
if err != nil {
	panic(err)
}
fmt.Println(ids)
```