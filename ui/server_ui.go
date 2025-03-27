package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type metadataPair struct {
	keyEntry   *widget.Entry
	valueEntry *widget.Entry
}

type connectionDetails struct {
	hostname string
	port     string
	metadata map[string]string
}

var metadata []*metadataPair

func (toolUI *UI) showServerUI(connDetails *connectionDetails) {
	if toolUI.ServerContent != nil && connDetails == nil {
		toolUI.ServerContent.Show()
		toolUI.CurrentView = ServerView
		return
	} else if toolUI.ServerContent != nil && connDetails != nil {
		toolUI.MainContent.Remove(toolUI.ServerContent)
	}
	toolUI.clearMetadata()

	serverBox := container.New(layout.NewVBoxLayout())

	serverForm := container.New(layout.NewFormLayout())

	hostLabel := widget.NewLabel("GRPC Server Hostname")
	hostLabel.Alignment = fyne.TextAlignTrailing
	serverForm.Add(hostLabel)

	hostEntry := widget.NewEntry()
	serverForm.Add(hostEntry)

	portLabel := widget.NewLabel("GRPC Server Port")
	portLabel.Alignment = fyne.TextAlignTrailing
	serverForm.Add(portLabel)

	portEntry := widget.NewEntry()
	serverForm.Add(portEntry)

	serverBox.Add(serverForm)

	sep := widget.NewSeparator()
	serverBox.Add(sep)

	meta := widget.NewLabel("Metadata")
	meta.Alignment = fyne.TextAlignCenter
	meta.TextStyle = fyne.TextStyle{Bold: true}
	meta.Wrapping = fyne.TextWrapWord
	serverBox.Add(meta)

	metaGrid := container.New(layout.NewGridLayout(2))
	serverBox.Add(metaGrid)

	addMetaDataButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		mke := widget.NewEntry()
		mke.SetPlaceHolder("Metadata Key")
		keyLabel := widget.NewLabel("Key")
		keyLabel.TextStyle = fyne.TextStyle{Bold: true}
		keyItem := container.New(layout.NewBorderLayout(nil, nil, keyLabel, nil), keyLabel, mke)

		mve := widget.NewEntry()
		mve.SetPlaceHolder("Metadata Value")
		valueLabel := widget.NewLabel("Value")
		valueLabel.TextStyle = fyne.TextStyle{Bold: true}
		valueItem := container.New(layout.NewBorderLayout(nil, nil, valueLabel, nil), valueLabel, mve)

		metaGrid.Add(keyItem)
		metaGrid.Add(valueItem)

		metadata = append(metadata, &metadataPair{mke, mve})
	})
	clearMetaDataButton := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		metaGrid.RemoveAll()
		metadata = nil
	})
	metadataButtonBox := container.New(layout.NewHBoxLayout(), addMetaDataButton, clearMetaDataButton)
	serverBox.Add(container.New(layout.NewBorderLayout(nil, nil, nil, metadataButtonBox),
		metadataButtonBox))

	submitButton := widget.NewButton("Submit", func() {
		var metaMap = make(map[string]string)
		for _, pair := range metadata {
			if pair.keyEntry.Text != "" && pair.valueEntry.Text != "" {
				metaMap[pair.keyEntry.Text] = pair.valueEntry.Text
			}
		}
		grpcConn.SetConnectionDetails(hostEntry.Text, portEntry.Text, metaMap)
		if hostEntry.Text == "" && portEntry.Text == "" {
			return
		}

		toolUI.ServerLabel.SetText(hostEntry.Text)
		toolUI.hideOrClearAllMainContent()
		toolUI.showProtoUI()
	})
	submitButton.Importance = widget.HighImportance
	submitButton.SetIcon(theme.ConfirmIcon())

	buttonBox := container.New(layout.NewBorderLayout(nil, nil, nil, submitButton), submitButton)
	buttonBox.Add(submitButton)
	serverBox.Add(buttonBox)

	if connDetails != nil {
		hostEntry.SetText(connDetails.hostname)
		portEntry.SetText(connDetails.port)
		metaGrid.RemoveAll()
		toolUI.clearMetadata()
		for key, value := range connDetails.metadata {
			mke := widget.NewEntry()
			mke.SetText(key)
			mke.SetPlaceHolder("Metadata Key")
			keyLabel := widget.NewLabel("Key")
			keyLabel.TextStyle = fyne.TextStyle{Bold: true}
			keyItem := container.New(layout.NewBorderLayout(nil, nil, keyLabel, nil), keyLabel, mke)

			mve := widget.NewEntry()
			mve.SetText(value)
			mve.SetPlaceHolder("Metadata Value")
			valueLabel := widget.NewLabel("Value")
			valueLabel.TextStyle = fyne.TextStyle{Bold: true}
			valueItem := container.New(layout.NewBorderLayout(nil, nil, valueLabel, nil), valueLabel, mve)

			metaGrid.Add(keyItem)
			metaGrid.Add(valueItem)

			metadata = append(metadata, &metadataPair{mke, mve})
		}
	}

	toolUI.ServerContent = container.NewScroll(serverBox)
	toolUI.ServerContent.ScrollToTop()
	toolUI.MainContent.Add(toolUI.ServerContent)
	toolUI.CurrentView = ServerView
}

func (toolUI *UI) clearMetadata() {
	metadata = nil
}
