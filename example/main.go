package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/runningmaster/smsclub"
)

const (
	apiBalance = "balance"
	apiSend    = "send"
	apiStatus  = "status"
)

func main() {
	flagCmd := flag.String("cmd", "", "v")
	flagUser := flag.String("user", "", "v")
	flagToken := flag.String("token", "", "v")
	flagSender := flag.String("sender", "", "v")
	flagText := flag.String("text", "", "v")
	flagListTo := flag.String("to", "", "csv")
	flagListID := flag.String("id", "", "csv")

	flag.Parse()

	sms, err := smsclub.New(
		smsclub.User(*flagUser),
		smsclub.Token(*flagToken),
		smsclub.Sender(*flagSender),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		flag.Usage()
		return
	}

	switch *flagCmd {
	case apiBalance:
		err = runBalance(sms)
	case apiSend:
		err = runSend(sms, *flagText, strings.Split(*flagListTo, ",")...)
	case apiStatus:
		err = runStatus(sms, strings.Split(*flagListID, ",")...)
	default:
		err = fmt.Errorf("unknown subcommand")
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
	}
}

func runBalance(sms smsclub.SMSCluber) error {
	bln, cre, err := sms.Balance()
	if err != nil {
		return err
	}

	fmt.Printf("balance: %.2f\ncredit: %.2f\n", bln, cre)
	return nil
}

func runSend(sms smsclub.SMSCluber, text string, to ...string) error {
	res, err := sms.Send(text, to...)
	if err != nil {
		return err
	}

	for i := range res {
		fmt.Println(res[i])
	}
	return nil
}

func runStatus(sms smsclub.SMSCluber, id ...string) error {
	res, err := sms.Status(id...)
	if err != nil {
		return err
	}

	for i := range res {
		fmt.Println(res[i])
	}
	return nil
}
