package main

import (
	"bufio"
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
	"os"
	"strconv"
	"strings"
)

func main() {
	os.Setenv("FYNE_FONT", "C:\\Windows\\Fonts\\msyh.ttc")
	defer os.Unsetenv("FYNE_FONT")
	guiApp := app.New()

	w := guiApp.NewWindow("U2 Key Reset Tool")
	w.CenterOnScreen()
	w.Resize(fyne.NewSize(800, 600))
	w.SetFixedSize(true)
	w.SetContent(fyne.NewContainer(makeForm(w, &guiApp)))
	w.SetOnClosed(func() {
		os.Exit(0)
	})

	w.ShowAndRun()
}

func makeLogWin(win *fyne.Window, guiApp *fyne.App) (*fyne.Window, *widget.Entry) {
	logWin := (*guiApp).NewWindow("U2 Key Reset Tool - Log")
	logWin.CenterOnScreen()
	logWin.Resize(fyne.NewSize(800, 600))
	logWin.SetFixedSize(true)
	logWin.SetOnClosed(func() {
		(*win).Show()
	})

	logEntity := widget.NewMultiLineEntry()
	logEntity.Resize(fyne.NewSize(350, 550))
	logEntity.Wrapping = fyne.TextWrapWord
	logEntityScroller := widget.NewVScrollContainer(logEntity)
	logEntityScroller.Resize(fyne.NewSize(350, 550))
	logWin.SetContent(fyne.NewContainerWithLayout(layout.NewMaxLayout(), logEntityScroller))

	return &logWin, logEntity
}

func makeForm(win fyne.Window, guiApp *fyne.App) fyne.CanvasObject {
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
	}

	container := fyne.NewContainerWithLayout(layout.NewMaxLayout(), form)

	form.OnSubmit = func() {
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

		logWin, logEntity := makeLogWin(&win, guiApp)
		(*logWin).Show()
		win.Hide()

		logEntity.SetText(logEntity.Text + "Submitted\n")

		old := os.Stdout // keep backup of the real stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// copy the output in a separate goroutine so printing can't block indefinitely
		go func() {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				logEntity.SetText(logEntity.Text + "\n" + scanner.Text())
			}
		}()

		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Error while resetting key!")
				fmt.Println(err)
				dialog.ShowError(errors.New("error while resetting key, see log for details"), *logWin)
			}
		}()
		doReset(config)

		// back to normal state
		w.Close()
		os.Stdout = old // restoring the real stdout
		dialog.ShowInformation("Success", "Key reset finished.", *logWin)
	}
	form.Resize(fyne.NewSize(350, 550))
	container.Resize(fyne.NewSize(800, 600))
	container.Refresh()
	return container
}

func doReset(config *u2.Config) {
	tool.InitClient(config)

	tool.TurnOnSilentMode()
	tool.ProcessTorrent()
}
