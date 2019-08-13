package main

import (
	"encoding/json"
	"fmt"
	"github.com/frikky/hive4go"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/samples/flags"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// Returns thehive login info
func getHiveLogin() (string, string, error) {
	var err error
	var configpath string

	configpath = "config.json"

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

	log.Printf("%#v", parsedRet)

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
	holder.AddPanel(textBox, "CaseID or Case Number")

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

		if len(word) == 0 {
			log.Println("Can't be empty!")
			os.Exit(3)
		}

		url, apikey, err := getHiveLogin()
		if err != nil {
			log.Printf("Credential error: %s", err)
			os.Exit(3)
		}

		hive := thehive.CreateLogin(url, apikey, false)

		// Try to get the ID
		if len(word) != 20 {
			searchJson := fmt.Sprintf(`{"query": {"_and": [{"_in": {"_field": "caseId", "_values": ["%s"]}}]}}`, word)

			ret, err := hive.FindCases([]byte(searchJson))
			if err != nil {
				log.Printf("Error finding case %s. Ret: %s\n", err, ret)
				os.Exit(3)
			}

			if strings.Contains(string(ret.Raw), "Invalid search query") {
				log.Printf("Error - CaseId and number %s doesn't exist.", word)
				os.Exit(3)
			}
			log.Printf(string(ret.Raw))
			GeneratePdf(hive, &ret.Detail[0])
		} else {
			ret, err := hive.GetCase(word)
			if err != nil {
				log.Println(err)
				os.Exit(3)
			}

			GeneratePdf(hive, ret)
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
