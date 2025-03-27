package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"grpc_ui_tool/proto"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// View is a constant for the different views supported in the main window
type View int

const (
	ServerView View = iota
	ProtoView
	InputView
)

// UI holds the main window widgets used across all views
type UI struct {
	App    fyne.App
	Window fyne.Window

	TopBorder  *fyne.Container
	MainBorder *fyne.Container

	TopLeft     *fyne.Container
	TopRight    *fyne.Container
	MainContent *fyne.Container

	ServerLabel *widget.Label

	HomeButton *widget.Button
	BackButton *widget.Button

	OpenButton *widget.Button
	SaveButton *widget.Button

	ServerContent *container.Scroll
	ProtoContent  *container.Scroll
	InputContent  *container.Scroll

	CurrentView View
}

var grpcConn *proto.GrpcConnection

// CreateUI creates the main window contents and takes a GrpcConnection to be used for getting values to show
func CreateUI(conn *proto.GrpcConnection) *UI {
	toolUI := &UI{}
	grpcConn = conn

	toolUI.App = app.New()
	toolUI.Window = toolUI.App.NewWindow("GRPC Tool")

	toolUI.HomeButton = widget.NewButtonWithIcon("", theme.HomeIcon(), func() {
		clearRequestStructure()
		toolUI.hideOrClearAllMainContent()
		toolUI.showServerUI(nil)
		toolUI.MainContent.Refresh()
	})

	toolUI.BackButton = widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		clearRequestStructure()
		toolUI.hideOrClearAllMainContent()
		switch toolUI.CurrentView {
		case ServerView:
			toolUI.showServerUI(nil)
		case ProtoView:
			toolUI.showServerUI(nil)
		case InputView:
			toolUI.showProtoUI()
		}
		toolUI.MainContent.Refresh()
	})

	toolUI.OpenButton = widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		openDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader == nil {
				return
			}
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
				return
			}

			var hostname, port string
			fileMetadata := make(map[string]string)

			file, err := os.Open(reader.URI().Path())
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
				return
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				split := strings.Split(line, ":")
				if len(split) < 2 {
					dialog.ShowError(fmt.Errorf("invalid line in server configuration: %s", line), toolUI.Window)
					return
				}
				switch split[0] {
				case "Hostname":
					hostname = split[1]
				case "Port":
					port = split[1]
				case "Metadata":
					if len(split) < 3 {
						dialog.ShowError(fmt.Errorf("invalid line in server configuration: %s", line), toolUI.Window)
						return
					}
					fileMetadata[split[1]] = split[2]
				}
			}
			err = file.Close()
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
				return
			}
			toolUI.showServerUI(&connectionDetails{
				hostname: hostname,
				port:     port,
				metadata: fileMetadata})
		}, toolUI.Window)
		openDialog.SetFilter(storage.NewExtensionFileFilter([]string{".gtserver"}))
		openDialog.SetView(dialog.ListView)
		openDialog.Show()
	})

	toolUI.SaveButton = widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		saveDialog := dialog.NewFileSave(func(closer fyne.URIWriteCloser, err error) {
			if closer == nil {
				return
			}
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
				return
			}
			serverConfig := "Hostname" + ":" + grpcConn.Hostname + "\n"
			serverConfig += "Port" + ":" + grpcConn.Port + "\n"
			for key, value := range grpcConn.Metadata {
				serverConfig += "Metadata:" + key + ":" + value + "\n"
			}
			_, err = closer.Write([]byte(serverConfig))
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
				return
			}
			err = closer.Close()
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
				return
			}
		}, toolUI.Window)
		saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".gtserver"}))
		saveDialog.SetFileName("Please Save With .gtserver Extension")
		saveDialog.SetView(dialog.ListView)
		saveDialog.Show()
	})

	toolUI.ServerLabel = widget.NewLabel("Server")
	toolUI.ServerLabel.Alignment = fyne.TextAlignCenter
	toolUI.ServerLabel.TextStyle = fyne.TextStyle{Bold: true}
	toolUI.ServerLabel.Wrapping = fyne.TextWrapWord

	toolUI.TopLeft = container.New(layout.NewHBoxLayout())
	toolUI.TopLeft.Add(toolUI.HomeButton)
	toolUI.TopLeft.Add(toolUI.BackButton)

	toolUI.TopRight = container.New(layout.NewHBoxLayout())
	toolUI.TopRight.Add(toolUI.OpenButton)
	toolUI.TopRight.Add(toolUI.SaveButton)

	toolUI.TopBorder = container.New(layout.NewBorderLayout(nil, nil,
		toolUI.TopLeft, toolUI.TopRight), toolUI.TopLeft, toolUI.TopRight, toolUI.ServerLabel)

	toolUI.MainContent = container.New(layout.NewStackLayout())

	toolUI.MainBorder = container.New(layout.NewBorderLayout(toolUI.TopBorder, nil,
		nil, nil), toolUI.TopBorder, toolUI.MainContent)

	toolUI.showServerUI(nil)

	toolUI.Window.SetContent(toolUI.MainBorder)
	toolUI.Window.Resize(fyne.NewSize(600, 400))
	toolUI.Window.ShowAndRun()

	return toolUI
}

func (toolUI *UI) hideOrClearAllMainContent() {
	if toolUI.ServerContent != nil {
		toolUI.ServerContent.Hide()
	}
	if toolUI.ProtoContent != nil {
		toolUI.ProtoContent.Hide()
	}
	if toolUI.InputContent != nil {
		toolUI.MainContent.Remove(toolUI.InputContent)
		toolUI.MainContent.Refresh()
	}
}
