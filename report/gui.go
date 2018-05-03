package main

import (
	"encoding/json"
	"fmt"
	"github.com/frikky/hive4go"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/samples/flags"
	"io/ioutil"
	"os"
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
	holder := theme.CreatePanelHolder()

	// Create temporary text in boxes
	textBox := theme.CreateTextBox()
	//textBox.SetText("HELO")

	// Add textboxes to panels
	holder.AddPanel(textBox, "CaseID")

	// Add panels
	splitterAB.AddChild(holder)

	// Create button
	button := theme.CreateButton()
	button.SetText("Click here to generate case statistics")

	// Button for creating stuff
	button.OnClick(func(gxui.MouseEvent) {
		lineStart := textBox.LineStart(0)
		lineEnd := textBox.LineEnd(0)

		word := textBox.TextAt(lineStart, lineEnd)

		// Generate the report
		url, apikey, err := getHiveLogin()
		hive := thehive.CreateLogin(url, apikey, false)
		ret, err := hive.GetCase(word)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
		}

		GeneratePdf(hive, ret)
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
