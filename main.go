package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	_ "github.com/i0range/U2KeyResetTool/driver/deluge"
	_ "github.com/i0range/U2KeyResetTool/driver/qBittorrent"
	_ "github.com/i0range/U2KeyResetTool/driver/transmission"
	"github.com/i0range/U2KeyResetTool/tool"
	"github.com/i0range/U2KeyResetTool/u2"
	"strconv"
	"strings"
)

func main() {
	guiApp := app.New()

	w := guiApp.NewWindow("U2 Key Reset Tool")
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)
	w.SetContent(fyne.NewContainer(makeForm(w)))

	w.ShowAndRun()
}

func makeForm(win fyne.Window) fyne.CanvasObject {
	target := widget.NewSelectEntry([]string{
		"Transmission",
		"qBittorrent",
		"Deluge",
	})
	host := widget.NewEntry()
	host.SetPlaceHolder("127.0.0.1")
	port := widget.NewEntry()
	port.SetPlaceHolder("9091")
	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Password")
	apiKey := widget.NewPasswordEntry()
	apiKey.SetPlaceHolder("API Key")
	proxy := widget.NewEntry()
	proxy.SetPlaceHolder("http://127.0.0.1:1080")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Target", Widget: target},
			{Text: "Host", Widget: host},
			{Text: "Port", Widget: port},
			{Text: "Username", Widget: username},
			{Text: "Password", Widget: password},
			{Text: "API Key", Widget: apiKey},
			{Text: "HTTPS Proxy", Widget: proxy},
		},
		OnSubmit: func() {
			fmt.Println("Form submitted")
			if target.Text == "" {
				dialog.ShowError(errors.New("please select target"), win)
				return
			}
			if host.Text == "" {
				dialog.ShowError(errors.New("host cannot be empty"), win)
				return
			}
			if port.Text == "" {
				dialog.ShowError(errors.New("port cannot be empty"), win)
				return
			}
			portInt, err := strconv.Atoi(port.Text)
			if err != nil {
				dialog.ShowError(err, win)
				return
			}
			if apiKey.Text == "" {
				dialog.ShowError(errors.New("API Key cannot be empty"), win)
				return
			}

			config := &u2.Config{
				Target: strings.TrimSpace(target.Text),
				Host:   strings.TrimSpace(host.Text),
				Port:   uint16(portInt),
				Secure: false,
				User:   strings.TrimSpace(username.Text),
				Pass:   strings.TrimSpace(password.Text),
				ApiKey: strings.TrimSpace(apiKey.Text),
				Proxy:  strings.TrimSpace(proxy.Text),
			}
			if config.Target == "Transmission" {
				config.Target = "transmission"
			} else if config.Target == "Deluge" {
				config.Target = "deluge"
			}

			defer func() {
				if err := recover(); err != nil {
					fmt.Println("Error while resetting key!")
					fmt.Println(err)
					dialog.ShowError(errors.New("error while resetting key, see log for details"), win)
				}
			}()
			doReset(config)
		},
	}

	return fyne.NewContainerWithLayout(layout.NewMaxLayout(), form)
}

func doReset(config *u2.Config) {
	tool.InitClient(config)

	tool.TurnOnSilentMode()
	tool.ProcessTorrent()
}
