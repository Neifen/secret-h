package game

import (
	"fmt"
	"github.com/skip2/go-qrcode"
	"net/http"
)

func CreateQr(gid string, req *http.Request) ([]byte, error) {
	host := req.Host
	pre := "https"
	if req.TLS == nil {
		pre = "http"
	}
	url := fmt.Sprintf("%s://%s/join-qr/%s", pre, host, gid)
	//fmt.Printf("QR URL: %s\n", url)
	return qrcode.Encode(url, qrcode.Medium, 256)
}
