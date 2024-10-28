package main

import (
	"context"
	"fmt"
)

type Message struct {
	Process func(string)
}

type MessageChannel chan *Message

func StartListener(ctx context.Context, messageChannel MessageChannel) error {
	voucherClient, err := InitVoucherClient(UnifiBaseURL, UnifiUser, UnifiPassword)
	if err != nil {
		return err
	}
	listen(ctx, messageChannel, voucherClient)
	return nil
}

func listen(ctx context.Context, messageChannel MessageChannel, voucherClient *Unifi) {
	for {
		select {
		case msg := <-messageChannel:
			logger.Info("Voucher Request")
			vouchers, err := voucherClient.GetVouchers()
			if err != nil {
				msg.Process(err.Error())
			} else if len(vouchers) == 0 {
				msg.Process("No vouchers found")
			} else {
				// Parse Voucher for code only
				// Return 5 vouchers
				var vouchersString string
				count := 5
				if len(vouchers) < count {
					count = len(vouchers)
				}
				for _, voucher := range vouchers[:count] {
					vouchersString = fmt.Sprintf("%s\n%s", voucher.Code, vouchersString)
				}
				msg.Process(vouchersString)
				logger.Info(fmt.Sprintf("Voucher retrieved: %s", vouchersString))
			}
		case <-ctx.Done():
			return
		}
	}
}
