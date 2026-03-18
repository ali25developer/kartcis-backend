package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	_ "github.com/lib/pq"
)

var WAClient *whatsmeow.Client
var WAQRCode string // Menyimpan QR code terbaru untuk discan

func InitWA() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost user=postgres password=password dbname=kartcis port=5432 sslmode=disable"
	}

	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New(context.Background(), "postgres", dbURL, dbLog)
	if err != nil {
		fmt.Println("Gagal inisialisasi WA storage:", err)
		return
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	WAClient = whatsmeow.NewClient(deviceStore, clientLog)
	WAClient.AddEventHandler(handler)

	if WAClient.Store.ID == nil {
		// Belum login, butuh QR
		qrChan, _ := WAClient.GetQRChannel(context.Background())
		err = WAClient.Connect()
		if err != nil {
			fmt.Println("Gagal connect WA:", err)
			return
		}
		go func() {
			for evt := range qrChan {
				if evt.Event == "code" {
					WAQRCode = evt.Code
					fmt.Println("WA QR Code updated. Silakan scan di dashboard admin.")

					// Optional: Save QR to file for debug
					qrcode.WriteFile(evt.Code, qrcode.Medium, 256, "uploads/wa_qr.png")
				} else {
					fmt.Println("WA Login event:", evt.Event)
				}
			}
		}()
	} else {
		// Sudah ada session, langsung connect
		err = WAClient.Connect()
		if err != nil {
			fmt.Println("Gagal connect WA:", err)
		} else {
			fmt.Println("WhatsApp connected!")
		}
	}
}

func handler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		fmt.Printf("Received a message from %s: %s\n", v.Info.Sender, v.Message.GetConversation())
	}
}

func SendWAMessage(target string, message string) error {
	if WAClient == nil || !WAClient.IsConnected() {
		return fmt.Errorf("WhatsApp tidak terkoneksi")
	}

	// Format nomor: 0812... -> 62812...
	if len(target) > 0 && target[0] == '0' {
		target = "62" + target[1:]
	}

	jid := types.NewJID(target, types.DefaultUserServer)

	msg := &waE2E.Message{
		Conversation: proto.String(message),
	}

	_, err := WAClient.SendMessage(context.Background(), jid, msg)
	return err
}
