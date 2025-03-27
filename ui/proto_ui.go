package ui

import (
	"bufio"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (toolUI *UI) showProtoUI() {
	if toolUI.ProtoContent != nil {
		toolUI.ProtoContent.Show()
		toolUI.CurrentView = ProtoView
		return
	}

	var protoButton *widget.Button
	var protoFile string
	var importEntries []*widget.Entry
	var importBox *fyne.Container

	protoBox := container.NewVBox()

	protoButton = widget.NewButton("Choose Proto File", func() {
		fileChoose := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if reader != nil {
				protoButton.SetText(reader.URI().Name())
				protoFile = reader.URI().Path()
				err = reader.Close()
				if err != nil {
					dialog.ShowError(err, toolUI.Window)
					return
				}
				importBox.RemoveAll()
				importEntries = nil

				protoDirectory := filepath.Dir(protoFile)
				importsFile := protoDirectory + "/imports.gtimport"
				file, err := os.Open(importsFile)
				if err != nil {
					dialog.ShowError(err, toolUI.Window)
					return
				}
				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					line := scanner.Text()
					impItem, imp := toolUI.getImportItem(line)
					importBox.Add(impItem)
					importEntries = append(importEntries, imp)
				}

				err = file.Close()
				if err != nil {
					dialog.ShowError(err, toolUI.Window)
					return
				}
			}
		}, toolUI.Window)
		fileChoose.SetFilter(storage.NewExtensionFileFilter([]string{".proto"}))
		fileChoose.SetView(dialog.ListView)
		fileChoose.Show()
	})
	protoLabel := widget.NewLabel("Protobuf File")
	protoLabel.TextStyle = fyne.TextStyle{Bold: true}
	protoButtonItem := container.New(layout.NewBorderLayout(nil, nil, protoLabel, nil), protoLabel, protoButton)
	protoBox.Add(protoButtonItem)

	sep := widget.NewSeparator()
	protoBox.Add(sep)

	imports := widget.NewLabel("Imports")
	imports.Alignment = fyne.TextAlignCenter
	imports.TextStyle = fyne.TextStyle{Bold: true}
	imports.Wrapping = fyne.TextWrapWord
	protoBox.Add(imports)

	importBox = container.New(layout.NewVBoxLayout())
	protoBox.Add(importBox)

	addImportButton := widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		impItem, imp := toolUI.getImportItem("")
		importBox.Add(impItem)
		importEntries = append(importEntries, imp)
	})
	clearImportsButton := widget.NewButtonWithIcon("", theme.ContentClearIcon(), func() {
		importBox.RemoveAll()
		importEntries = nil
	})
	activity := widget.NewActivity()
	activity.Hide()
	saveImportsButton := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		if protoFile == "" {
			return
		}
		activity.Show()
		activity.Start()
		protoDirectory := filepath.Dir(protoFile)
		saveFile := protoDirectory + "/imports.gtimport"
		if importEntries != nil {
			fileOutput := ""
			for _, imp := range importEntries {
				fileOutput += imp.Text + "\n"
			}
			err := os.WriteFile(saveFile, []byte(fileOutput), 0777)
			if err != nil {
				dialog.ShowError(err, toolUI.Window)
			}
		}
		activity.Stop()
		activity.Hide()
	})
	stack := container.NewStack(saveImportsButton, activity)
	importButtonBox := container.New(layout.NewHBoxLayout(), addImportButton, clearImportsButton, stack)
	protoBox.Add(container.New(layout.NewBorderLayout(nil, nil, nil, importButtonBox),
		importButtonBox))

	submitButton := widget.NewButton("Submit", func() {
		if protoFile == "" {
			return
		}
		var importPaths []string
		for _, ent := range importEntries {
			if ent.Text != "" {
				importPaths = append(importPaths, ent.Text)
			}
		}
		err := grpcConn.LoadRegistry(importPaths, protoFile)
		if err != nil {
			dialog.ShowError(err, toolUI.Window)
			return
		}
		toolUI.hideOrClearAllMainContent()
		toolUI.showInputUI()
	})
	submitButton.Importance = widget.HighImportance
	submitButton.SetIcon(theme.ConfirmIcon())

	buttonBox := container.New(layout.NewBorderLayout(nil, nil, nil, submitButton), submitButton)
	buttonBox.Add(submitButton)
	protoBox.Add(buttonBox)

	toolUI.ProtoContent = container.NewScroll(protoBox)
	toolUI.ProtoContent.ScrollToTop()
	toolUI.MainContent.Add(toolUI.ProtoContent)
	toolUI.CurrentView = ProtoView
}

func (toolUI *UI) getImportItem(importPath string) (*fyne.Container, *widget.Entry) {
	imp := widget.NewEntry()
	imp.SetPlaceHolder("Import Path")
	if importPath != "" {
		imp.SetText(importPath)
	}
	impLabel := widget.NewLabel("Path")
	impLabel.TextStyle = fyne.TextStyle{Bold: true}
	impItem := container.New(layout.NewBorderLayout(nil, nil, impLabel, nil), impLabel, imp)
	return impItem, imp
}
