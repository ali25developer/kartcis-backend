package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kartcis-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

// GET /admin/broadcast/wa/qr
func GetWAStatus(c *gin.Context) {
	connected := false
	if utils.WAClient != nil {
		connected = utils.WAClient.IsConnected() && utils.WAClient.IsLoggedIn()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"connected": connected,
			"qr_code":   utils.WAQRCode,
		},
	})
}

// POST /admin/broadcast/wa/send
func BroadcastWA(c *gin.Context) {
	if utils.WAClient == nil || !utils.WAClient.IsConnected() || !utils.WAClient.IsLoggedIn() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "WhatsApp belum terhubung."})
		return
	}

	messageTemplate := c.PostForm("message")
	file, err := c.FormFile("file")
	if err != nil || messageTemplate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Pesan dan file Excel diperlukan"})
		return
	}

	tempPath := filepath.Join("uploads", fmt.Sprintf("broadcast_%d_%s", time.Now().Unix(), file.Filename))
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal simpan file"})
		return
	}

	// Langsung jawab ke frontend agar browser tidak timeout
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Broadcast sedang berjalan di latar belakang. Estimasi waktu: ~30-60 menit untuk 1000 pesan.",
	})

	// Jalankan proses pengiriman di background (Goroutine)
	go func() {
		defer os.Remove(tempPath)

		f, err := excelize.OpenFile(tempPath)
		if err != nil {
			fmt.Println("Error buka excel:", err)
			return
		}
		defer f.Close()

		rows, _ := f.GetRows("Sheet1")
		for i, row := range rows {
			if i == 0 || len(row) < 2 {
				continue
			}

			name := row[0]
			phone := strings.TrimSpace(row[1])
			phone = strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, phone)

			if phone == "" {
				continue
			}

			msg := strings.ReplaceAll(messageTemplate, "{nama}", name)
			_ = utils.SendWAMessage(phone, msg)

			// Jeda acak 3 - 7 detik agar lebih aman dari BAN
			randomDelay := rand.Intn(4) + 3
			time.Sleep(time.Duration(randomDelay) * time.Second)
		}
		fmt.Println("Broadcast WhatsApp selesai!")
	}()
}
