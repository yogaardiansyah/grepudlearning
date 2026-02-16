package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendVerificationEmail(toEmail, code string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	
	senderName := os.Getenv("SMTP_SENDER_NAME")
	if senderName == "" {
		senderName = user
	}

	subject := "Subject: Kode Verifikasi Food App\n"
	fromHeader := fmt.Sprintf("From: %s\n", senderName)
	toHeader := fmt.Sprintf("To: %s\n", toEmail)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; padding: 20px;">
				<div style="background-color: #f4f4f4; padding: 20px; border-radius: 8px;">
					<h2 style="color: #333;">Verifikasi Akun</h2>
					<p>Terima kasih telah mendaftar. Gunakan kode berikut:</p>
					<h1 style="color: #0070f3; background: #fff; padding: 10px; display: inline-block;">%s</h1>
					<p>Kode berlaku selama 15 menit.</p>
				</div>
			</body>
		</html>
	`, code)

	msg := []byte(subject + fromHeader + toHeader + mime + body)
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, password, host)

	return smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
}


func SendReceiptEmail(toEmail, username, orderID string, amount float64, itemName string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASS")
	senderName := os.Getenv("SMTP_SENDER_NAME")

	subject := fmt.Sprintf("Subject: Struk Pembayaran Order #%s\n", orderID)
	fromHeader := fmt.Sprintf("From: %s\n", senderName)
	toHeader := fmt.Sprintf("To: %s\n", toEmail)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	// Format Rupiah
	formattedAmount := fmt.Sprintf("Rp %.0f", amount)

	body := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; padding: 20px; background-color: #f4f4f4;">
				<div style="max-width: 600px; margin: auto; background: white; padding: 20px; border-radius: 8px; border: 1px solid #ddd;">
					<h2 style="color: #27ae60; text-align: center;">Pembayaran Berhasil!</h2>
					<p>Halo <strong>%s</strong>,</p>
					<p>Terima kasih telah melakukan pembayaran. Berikut detail pesanan Anda:</p>
					
					<table style="width: 100%%; border-collapse: collapse; margin-top: 20px;">
						<tr style="background: #eee;">
							<td style="padding: 10px;">No. Order</td>
							<td style="padding: 10px; font-weight: bold;">#%s</td>
						</tr>
						<tr>
							<td style="padding: 10px;">Menu</td>
							<td style="padding: 10px;">%s</td>
						</tr>
						<tr style="background: #eee;">
							<td style="padding: 10px;">Total Bayar</td>
							<td style="padding: 10px; font-weight: bold; color: #27ae60;">%s</td>
						</tr>
						<tr>
							<td style="padding: 10px;">Status</td>
							<td style="padding: 10px;"><span style="background: #27ae60; color: white; padding: 2px 6px; border-radius: 4px;">LUNAS</span></td>
						</tr>
					</table>

					<p style="margin-top: 20px; text-align: center; color: #888; font-size: 12px;">Simpan email ini sebagai bukti pembayaran yang sah.</p>
				</div>
			</body>
		</html>
	`, username, orderID, itemName, formattedAmount)

	msg := []byte(subject + fromHeader + toHeader + mime + body)
	addr := fmt.Sprintf("%s:%s", host, port)
	auth := smtp.PlainAuth("", user, password, host)

	return smtp.SendMail(addr, auth, user, []string{toEmail}, msg)
}