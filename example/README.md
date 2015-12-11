This is a small example of using that package in CLI app.

```
USAGE:
main <command> [--flag=<value>,...]
commands:
	balance - get balance (and credit)
	--user=string:string - username:password

	send - send SMS to recipients
	--user=string:string - username:password
	--from=string - alphaname
	--text=string - message of SMS
	--to=string,... - list of phone numbers (comma-separated)
	--lt=int - lifetime om SMS in minutes (default 0)

	status - get SMS status
	--user=string:string - username:password
	--id=string,... - list of SMS ID from 'send' command (comma-separated)

```