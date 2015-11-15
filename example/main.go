package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/runningmaster/smsclub"
)

const (
	apiBalance = "balance"
	apiSend    = "send"
	apiStatus  = "status"
)

var (
	apiUsage = map[string]string{
		apiBalance: "get balance (and credit)",
		apiSend:    "send SMS to recipients",
		apiStatus:  "get SMS status",
	}

	cmdBalance = &command{
		run:  runBalance,
		name: apiBalance,
		desc: apiUsage[apiBalance],
	}
	cmdSend = &command{
		run:  runSend,
		name: apiSend,
		desc: apiUsage[apiSend],
	}
	cmdStatus = &command{
		run:  runStatus,
		name: apiStatus,
		desc: apiUsage[apiStatus],
	}
	commands = []*command{
		cmdBalance,
		cmdSend,
		cmdStatus,
	}

	flagUser   string
	flagText   string
	flagFrom   string
	flagListTo string
	flagListID string
	flagTime   time.Duration

	sms smsclub.SMSer
)

type command struct {
	run   func() error
	name  string
	desc  string
	flags flag.FlagSet
}

func init() {
	flag.Usage = usage
	for _, cmd := range commands {
		cmd.flags.StringVar(&flagUser, "user", "", "--user=string:string - username:password")
		switch {
		case cmd.name == apiSend:
			cmd.flags.StringVar(&flagText, "text", "", "--text=string - message of SMS")
			cmd.flags.StringVar(&flagFrom, "from", "", "--from=string - alphaname")
			cmd.flags.StringVar(&flagListTo, "to", "", "--to=string,... - list of phone numbers (comma-separated)")
			cmd.flags.DurationVar(&flagTime, "lt", 0, "--lt=int - lifetime om SMS in minutes (default 0)")
		case cmd.name == apiStatus:
			cmd.flags.StringVar(&flagListID, "id", "", "--id=string,... - list of SMS ID from 'send' command (comma-separated)")
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()
	if len(args) < 1 || args[0] == "help" {
		usage()
	}

	var err error
	for _, cmd := range commands {
		if cmd.name == args[0] {
			err = cmd.flags.Parse(args[1:])
			if err != nil {
				log.Fatal(err)
			}
			sms = smsclub.New(splitUser(flagUser))
			err = cmd.run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
				usage()
			}
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown subcommand %q\n", args[0])
	usage()
}

func usage() {
	w := os.Stderr
	fmt.Fprintln(w, "USAGE:")
	fmt.Fprintln(w, appName(), "<command>", "[--flag=<value>,...]")
	fmt.Fprintln(w, "commands:")
	for _, cmd := range commands {
		fmt.Fprintf(w, "\t%s - %s\n", cmd.name, cmd.desc)
		fmt.Fprintf(w, "\t%s\n", cmd.flags.Lookup("user").Usage)
		switch {
		case cmd.name == apiSend:
			fmt.Fprintf(w, "\t%s\n", cmd.flags.Lookup("text").Usage)
			fmt.Fprintf(w, "\t%s\n", cmd.flags.Lookup("from").Usage)
			fmt.Fprintf(w, "\t%s\n", cmd.flags.Lookup("to").Usage)
			fmt.Fprintf(w, "\t%s\n", cmd.flags.Lookup("lt").Usage)
		case cmd.name == apiStatus:
			fmt.Fprintf(w, "\t%s\n", cmd.flags.Lookup("id").Usage)
		}
		fmt.Fprintln(w, "")
	}
	os.Exit(2)
}

func runBalance() error {
	bln, cre, err := sms.Balance()
	if err != nil {
		return err
	}

	fmt.Printf("balance: %.2f\ncredit: %.2f\n", bln, cre)
	return nil
}

func runSend() error {
	err := sms.LifeTime(flagTime)
	if err != nil {
		return err
	}

	res, err := sms.Send(flagText, flagFrom, strings.Split(flagListTo, ",")...)
	if err != nil {
		return err
	}

	for i := range res {
		fmt.Println(res[i])
	}
	return nil
}

func runStatus() error {
	res, err := sms.Status(strings.Split(flagListID, ",")...)
	if err != nil {
		return err
	}

	for i := range res {
		fmt.Println(res[i])
	}
	return nil
}

func splitUser(user string) (name, pass string) {
	namepass := strings.Split(user, ":")
	if len(namepass) == 2 {
		name = namepass[0]
		pass = namepass[1]
	}
	return name, pass
}

func appName() string {
	return filepath.Base(os.Args[0])
}
