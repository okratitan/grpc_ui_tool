package ui

import (
	"fmt"
	"image/color"

	"grpc_ui_tool/proto"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type inputWidgetType int

const (
	gridMessage inputWidgetType = iota
	gridOneOf
	entry
	sel
	check
)

type message struct {
	fields []*messageField
	offset float32
}

type messageField struct {
	parent    *message
	field     *proto.Field
	nested    *message
	inputType inputWidgetType
	grid      *fyne.Container
	entry     *widget.Entry
	sel       *widget.Select
	check     *widget.Check
}

var fieldStructure *message

func (toolUI *UI) showInputUI() {
	var serviceSelect, methodSelect *widget.Select

	fieldStructure = &message{}

	content := container.New(layout.NewVBoxLayout())

	inputBox := container.New(layout.NewVBoxLayout())
	inputGrid := container.New(layout.NewGridLayout(2))
	inputBox.Add(inputGrid)

	serviceMethodGrid := container.New(layout.NewGridLayout(2))

	serviceSelect = widget.NewSelect([]string{}, func(selected string) {
		methods, err := grpcConn.GetMethods(selected)
		if err != nil {
			dialog.ShowError(err, toolUI.Window)
			return
		}
		methodSelect.SetOptions(methods)
	})
	serviceSelectLabel := toolUI.getFieldLabel("Service")
	serviceSelectItem := container.New(layout.NewBorderLayout(nil, nil, serviceSelectLabel, nil), serviceSelectLabel, serviceSelect)
	services, err := grpcConn.GetServices()
	if err != nil {
		dialog.ShowError(err, toolUI.Window)
		return
	}
	serviceSelect.SetOptions(services)

	methodSelect = widget.NewSelect([]string{}, func(selected string) {
		fields, err := grpcConn.GetFields(serviceSelect.Selected+"."+methodSelect.Selected, proto.Input)
		if err != nil {
			dialog.ShowError(err, toolUI.Window)
			return
		}
		inputBox.RemoveAll()
		inputGrid = container.New(layout.NewGridLayout(2))
		inputBox.Add(inputGrid)
		fieldStructure = &message{}
		toolUI.createRequestStructure(fields, nil, inputBox, inputGrid)
	})
	methodSelectLabel := toolUI.getFieldLabel("Method")
	methodSelectItem := container.New(layout.NewBorderLayout(nil, nil, methodSelectLabel, nil), methodSelectLabel, methodSelect)

	serviceMethodGrid.Add(serviceSelectItem)
	serviceMethodGrid.Add(methodSelectItem)

	content.Add(serviceMethodGrid)
	content.Add(inputBox)

	activity := widget.NewActivity()
	activity.Hide()
	submitButton := widget.NewButton("Submit", func() {
		activity.Show()
		activity.Start()
		jsonString := toolUI.getRequestJson()

		resp, err := grpcConn.Send(serviceSelect.Selected, methodSelect.Selected, jsonString)
		if err != nil {
			dialog.ShowError(err, toolUI.Window)
			return
		}

		textGrid := widget.NewTextGridFromString(resp)

		dialog.ShowCustom("GRPC Response", "OK", textGrid, toolUI.Window)
		activity.Stop()
		activity.Hide()
	})
	submitButton.Importance = widget.HighImportance
	submitButton.SetIcon(theme.ConfirmIcon())

	stack := container.NewStack(submitButton, activity)

	buttonBox := container.New(layout.NewBorderLayout(nil, nil, nil, stack), stack)
	content.Add(buttonBox)

	toolUI.InputContent = container.NewScroll(content)
	toolUI.InputContent.ScrollToTop()
	toolUI.MainContent.Add(toolUI.InputContent)
	toolUI.CurrentView = InputView
}

func clearRequestStructure() {
	fieldStructure = nil
}

func (toolUI *UI) createRequestStructure(fields []*proto.Field, parent *message, inputBox *fyne.Container, inputGrid *fyne.Container) {
	for _, field := range fields {
		msgField := &messageField{
			field:  field,
			parent: parent,
		}
		if field.IsMessage {
			offset := float32(0)
			if parent != nil {
				offset += parent.offset
			}
			offset += 32

			msgField.nested = &message{offset: offset}
			msgField.inputType = gridMessage

			spacer := canvas.NewRectangle(color.Transparent)
			spacer.SetMinSize(fyne.NewSize(offset, 0))
			spacer.Show()
			msgField.grid = container.New(layout.NewGridLayout(2))
			cont := container.New(layout.NewBorderLayout(nil, nil, spacer, nil), spacer, msgField.grid)
			inputGrid.Add(toolUI.getFieldLabel(msgField.field.Name + " (" + msgField.field.Type + "):"))
			inputGrid.Add(widget.NewLabel(""))
			inputBox.Add(cont)

			toolUI.createRequestStructure(field.FieldMessage.Fields, msgField.nested, inputBox, msgField.grid)

			inputGrid = container.New(layout.NewGridLayout(2))
			inputBox.Add(inputGrid)
		} else if field.IsOneOf {
			msgField.nested = nil
			msgField.inputType = gridOneOf
			cont := container.New(layout.NewVBoxLayout())
			msgField.grid = container.New(layout.NewGridLayout(2))
			cont.Add(msgField.grid)
			msgField.sel = widget.NewSelect(field.FieldOneOf.OneOfKeys, func(selected string) {
				msgField.nested = nil
				msgField.nested = &message{}
				cont.RemoveAll()
				msgField.grid = container.New(layout.NewGridLayout(2))
				cont.Add(msgField.grid)

				toolUI.createRequestStructure(field.FieldOneOf.OneOfValues[selected], msgField.nested, cont, msgField.grid)

				inputGrid = container.New(layout.NewGridLayout(2))
				inputBox.Add(inputGrid)
			})
			inputGrid.Add(toolUI.getFieldLabel(msgField.field.Name + " (" + msgField.field.Type + "):"))
			inputGrid.Add(msgField.sel)
			inputBox.Add(cont)
		} else if field.IsEnum {
			msgField.inputType = sel
			var enums []string
			for _, val := range field.EnumValues {
				enums = append(enums, val.Name)
			}
			msgField.sel = widget.NewSelect(enums, nil)
			inputGrid.Add(toolUI.getFieldLabel(msgField.field.Name + " (" + msgField.field.Type + "):"))
			inputGrid.Add(msgField.sel)
		} else if field.Type == "bool" {
			msgField.inputType = check
			msgField.check = widget.NewCheck("", nil)
			inputGrid.Add(toolUI.getFieldLabel(msgField.field.Name + " (" + msgField.field.Type + "):"))
			inputGrid.Add(msgField.check)
		} else {
			msgField.inputType = entry
			msgField.entry = widget.NewEntry()
			inputGrid.Add(toolUI.getFieldLabel(msgField.field.Name + " (" + msgField.field.Type + "):"))
			inputGrid.Add(msgField.entry)
		}
		if parent != nil {
			parent.fields = append(parent.fields, msgField)
		} else {
			fieldStructure.fields = append(fieldStructure.fields, msgField)
		}
	}
}

func (toolUI *UI) getFieldLabel(text string) *widget.Label {
	label := widget.NewLabel(text)
	label.TextStyle = fyne.TextStyle{Bold: true}
	return label
}

func getNestedJson(msg *message) string {
	msgFirst := true
	jsonString := ""
	for _, m := range msg.fields {
		if m.nested == nil && m.field.IsOneOf {
			continue
		}
		if !msgFirst {
			jsonString += ", "
		}
		if m.nested != nil && m.field.IsMessage {
			jsonString += "\"" + m.field.JsonName + "\": { "
			jsonString += getNestedJson(m.nested)
			jsonString += " }"
		} else if m.nested != nil && m.field.IsOneOf {
			jsonString += getNestedJson(m.nested)
		} else {
			jsonString += "\"" + m.field.JsonName + "\": "
			switch m.inputType {
			case entry:
				jsonString += "\"" + m.entry.Text + "\""
			case check:
				if m.check.Checked {
					jsonString += "true"
				} else {
					jsonString += "false"
				}
			case sel:
				jsonString += "\"" + m.sel.Selected + "\""
			default:
				continue
			}
		}
		msgFirst = false
	}
	return jsonString
}

func (toolUI *UI) getRequestJson() string {
	jsonString := "{ "
	jsonString += getNestedJson(fieldStructure)
	jsonString += " }"
	return jsonString
}

// For Debug Printing the JSON
func (toolUI *UI) printRequestStructure() {
	jsonString := "{ "
	jsonString += getNestedJson(fieldStructure)
	jsonString += " }"
	fmt.Println(jsonString)
}
