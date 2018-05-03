package main

import (
	"encoding/json"
	"fmt"
	"github.com/aodin/date"
	"github.com/frikky/hive4go"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/samples/flags"
	"io/ioutil"
	"os"
	"time"
)

// Returns thehive login info
func getHiveLogin() (string, string, error) {
	var err error
	var configpath string

	configpath = "../config.json"

	file, err := ioutil.ReadFile(configpath)

	if err != nil {
		fmt.Printf("Error getting hive json: %s\n", err)
		return "", "", err
	}

	type hive struct {
		HiveUrl string `json:"hiveurl"`
		HiveApi string `json:"hiveapikey"`
	}

	parsedRet := new(hive)
	err = json.Unmarshal(file, parsedRet)
	if err != nil {
		fmt.Printf("Error unmarshaling hive json: %s\n", err)
		return "", "", err
	}

	return parsedRet.HiveUrl, parsedRet.HiveApi, nil
}

func appMain(driver gxui.Driver) {
	theme := flags.CreateTheme(driver)

	// Create splitter for textboxes
	splitterAB := theme.CreateSplitterLayout()
	splitterAB.SetOrientation(gxui.Horizontal)

	// Create text holders
	holderFrom := theme.CreatePanelHolder()
	holderTo := theme.CreatePanelHolder()

	// Create temporary text in boxes
	text := date.New(time.Now().Year(), time.Now().Month(), time.Now().Day()).String()
	textBoxFrom := theme.CreateTextBox()
	textBoxFrom.SetText(text)
	textBoxTo := theme.CreateTextBox()
	textBoxTo.SetText(text)

	// Add textboxes to panels
	holderFrom.AddPanel(textBoxFrom, "From")
	holderTo.AddPanel(textBoxTo, "To")

	// Add panels
	splitterAB.AddChild(holderFrom)
	splitterAB.AddChild(holderTo)

	// Create button
	button := theme.CreateButton()
	button.SetText("Click here to generate case statistics")

	// Button for creating stuff
	button.OnClick(func(gxui.MouseEvent) {
		lineFromStart := textBoxFrom.LineStart(0)
		lineFromEnd := textBoxFrom.LineEnd(0)

		lineToStart := textBoxTo.LineStart(0)
		lineToEnd := textBoxTo.LineEnd(0)

		wordFrom := textBoxFrom.TextAt(lineFromStart, lineFromEnd)
		wordTo := textBoxTo.TextAt(lineToStart, lineToEnd)

		// Should always be 10 in length
		/*
			if len(wordFrom) != 10 {
				fmt.Println("INVALID STUFF IN FROM")
			} else if len(wordTo) != 10 {
				fmt.Println("INVALID STUFF IN TO")
			}
		*/

		// Generate the report
		url, apikey, err := getHiveLogin()
		hive := thehive.CreateLogin(url, apikey, false)

		err = getreport(hive, wordFrom, wordTo)
		if err != nil {
			os.Exit(3)
		}
	})

	// Split vertically
	vSplitter := theme.CreateSplitterLayout()
	vSplitter.AddChild(splitterAB)
	vSplitter.AddChild(button)

	// Generate the window
	window := theme.CreateWindow(400, 100, "Case statistics")
	window.SetScale(flags.DefaultScaleFactor)
	window.AddChild(vSplitter)
	window.OnClose(driver.Terminate)
}

func main() {
	gl.StartDriver(appMain)
}
