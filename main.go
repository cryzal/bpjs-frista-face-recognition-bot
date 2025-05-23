package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/lxn/win"
	"github.com/micmonay/keybd_event"
	"golang.org/x/sys/windows"
)

const (
	port            = ":3006"
	LoginWindowName = "Login Frista (Face Recognition BPJS Kesehatan)"
	MainWindowName  = "Frista (Face Recognition BPJS Kesehatan) 3.0.2"
	TimeoutDuration = 30 * time.Second
)

// Mapping karakter ke konstanta virtual keycode
var keyMap = map[rune]int{
	'1': keybd_event.VK_1, '2': keybd_event.VK_2, '3': keybd_event.VK_3,
	'4': keybd_event.VK_4, '5': keybd_event.VK_5, '6': keybd_event.VK_6,
	'7': keybd_event.VK_7, '8': keybd_event.VK_8, '9': keybd_event.VK_9,
	'0': keybd_event.VK_0, 'q': keybd_event.VK_Q, 'w': keybd_event.VK_W,
	'e': keybd_event.VK_E, 'r': keybd_event.VK_R, 't': keybd_event.VK_T,
	'y': keybd_event.VK_Y, 'u': keybd_event.VK_U, 'i': keybd_event.VK_I,
	'o': keybd_event.VK_O, 'p': keybd_event.VK_P, 'a': keybd_event.VK_A,
	's': keybd_event.VK_S, 'd': keybd_event.VK_D, 'f': keybd_event.VK_F,
	'g': keybd_event.VK_G, 'h': keybd_event.VK_H, 'j': keybd_event.VK_J,
	'k': keybd_event.VK_K, 'l': keybd_event.VK_L, 'z': keybd_event.VK_Z,
	'x': keybd_event.VK_X, 'c': keybd_event.VK_C, 'v': keybd_event.VK_V,
	'b': keybd_event.VK_B, 'n': keybd_event.VK_N, 'm': keybd_event.VK_M,
	' ': keybd_event.VK_SPACE, '.': keybd_event.VK_DOT, ',': keybd_event.VK_COMMA,
	'-': keybd_event.VK_MINUS, '=': keybd_event.VK_EQUAL,
}

// Mapping karakter yang butuh SHIFT
var shiftKeyMap = map[rune]int{
	'!': keybd_event.VK_1, '@': keybd_event.VK_2, '#': keybd_event.VK_3,
	'$': keybd_event.VK_4, '%': keybd_event.VK_5, '^': keybd_event.VK_6,
	'&': keybd_event.VK_7, '*': keybd_event.VK_8, '(': keybd_event.VK_9,
	')': keybd_event.VK_0, '_': keybd_event.VK_MINUS, '+': keybd_event.VK_EQUAL,
}

type OpenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	NoBpjs   string `json:"no_bpjs"`
}

type CloseRequest struct {
	AppName string `json:"app_name"`
}

func refocusApplication(windowName string) {
	namePtr, err := windows.UTF16PtrFromString(windowName)
	if err != nil {
		fmt.Println("Error converting window name:", err)
		return
	}

	hwnd := win.FindWindow(nil, namePtr)
	if hwnd != 0 {
		win.ShowWindow(hwnd, win.SW_RESTORE) // Jika aplikasi minimize, restore
		win.SetForegroundWindow(hwnd)        // Bawa ke depan
	} else {
		fmt.Println("Aplikasi tidak ditemukan:", windowName)
	}
}

func main() {
	fmt.Println("Letakan file ini sejajar dengan instalan Frista (frista.exe)")

	app := fiber.New(fiber.Config{
		AppName: "Restapi Frista BPJS",
	})

	// Tambahkan middleware CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "*",
	}))

	app.Post("/ping", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"message": "READY!!!"})
	})

	app.Post("/open", func(c *fiber.Ctx) error {
		var request OpenRequest
		if err := c.BodyParser(&request); err != nil {
			return c.Status(422).JSON(fiber.Map{"message": "Bad Request"})
		}

		filePath := getExePath("frista.exe")
		cmd := exec.Command(filePath)
		if err := cmd.Start(); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": err.Error()})
		}

		// Tunggu aplikasi Login Frista terbuka
		if err := waitForWindow(LoginWindowName, 10*time.Second); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Aplikasi tidak terbuka dalam batas waktu"})
		}

		time.Sleep(500 * time.Millisecond)

		refocusApplication(LoginWindowName)

		// Input Username & Password

		typeStr(request.Username)
		pressTab()
		typeStr(request.Password)
		time.Sleep(500 * time.Millisecond)
		pressTab()
		pressSpace() // Tekan SPACE untuk login

		// Tunggu aplikasi utama Frista terbuka
		if err := waitForWindow(MainWindowName, 10*time.Second); err != nil {
			return c.Status(500).JSON(fiber.Map{"message": "Aplikasi tidak terbuka dalam batas waktu"})
		}
		time.Sleep(1000 * time.Millisecond)
		// Refocus ke aplikasi utama & input No BPJS
		refocusApplication(MainWindowName)
		typeStr(request.NoBpjs)

		return c.Status(200).JSON(fiber.Map{"message": "success"})
	})

	app.Post("/close", func(c *fiber.Ctx) error {

		filePath := getExePath("frista.exe")
		cmd := exec.Command("TASKKILL", "/IM", filepath.Base(filePath), "/F")
		cmd.Run()
		return c.Status(200).JSON(fiber.Map{"message": "success"})
	})

	err := app.Listen(port)
	if err != nil {
		fmt.Println("Error starting the server:", err)
	}
}

// Mendapatkan path file exe
func getExePath(app string) string {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return ""
	}
	return filepath.Join(currentDir, app)
}

// Menekan tombol TAB
func pressTab() error {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	kb.SetKeys(keybd_event.VK_TAB)
	err = kb.Launching()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// Menekan tombol SPACE
func pressSpace() error {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	kb.SetKeys(keybd_event.VK_SPACE)
	err = kb.Launching()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// Mengetik string secara otomatis
func typeStr(text string) {
	kb, err := keybd_event.NewKeyBonding()
	if err != nil {
		fmt.Println("Error initializing keyboard:", err)
		return
	}

	for _, char := range text {
		kb.Clear() // Reset keybinding sebelum setiap karakter

		// Cek apakah huruf besar
		if unicode.IsUpper(char) {
			kb.HasSHIFT(true)
			char = unicode.ToLower(char) // Gunakan versi lowercase agar sesuai keycode
		}

		// Cek apakah karakter butuh SHIFT
		if keyCode, exists := shiftKeyMap[char]; exists {
			kb.HasSHIFT(true)
			kb.SetKeys(keyCode)
		} else if keyCode, exists := keyMap[char]; exists {
			kb.SetKeys(keyCode)
		} else {
			fmt.Println("Unsupported character:", string(char))
			continue
		}

		// Tekan tombol
		err = kb.Launching()
		if err != nil {
			fmt.Println("Error typing character:", string(char), err)
		}

		// Delay agar tidak terlalu cepat
		//time.Sleep(50 * time.Millisecond)
	}
}

// Fungsi untuk menunggu aplikasi terbuka dalam batas waktu tertentu
func waitForWindow(windowName string, timeout time.Duration) error {
	startTime := time.Now()
	namePtr, _ := windows.UTF16PtrFromString(windowName)

	for {
		if win.FindWindow(nil, namePtr) != 0 {
			return nil // Jika ditemukan, keluar dari fungsi
		}
		if time.Since(startTime) > timeout {
			return fmt.Errorf("window '%s' tidak ditemukan dalam batas waktu", windowName)
		}
		time.Sleep(500 * time.Millisecond) // Cek setiap 500ms
	}
}
