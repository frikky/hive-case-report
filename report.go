package main

import (
	"encoding/json"
	"fmt"
	"github.com/frikky/hive4go"
	"github.com/google/gxui"
	"github.com/google/gxui/drivers/gl"
	"github.com/google/gxui/samples/flags"
	"github.com/signintech/gopdf"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"time"
)

// Global pdf variable
var pdf gopdf.GoPdf

func getTime() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// Cleans up markdown stuff
func cleanupText(text string) string {
	textBytes := []byte(text)
	newBytes := []byte{}

	skip := false
	for bytecnt, curbyte := range textBytes {
		// Random chars
		if string(curbyte) == "\n" {
			newBytes = append(newBytes, byte(32))
		}

		if curbyte < 31 || curbyte > 126 {
			continue
		}

		// cba finding byte
		if string(textBytes[bytecnt]) == "&" && string(textBytes[bytecnt+3]) == ";" {
			skip = true
		}

		if !skip {
			newBytes = append(newBytes, curbyte)
		}

		if skip && string(curbyte) == ";" {
			skip = false
		}

	}
	return string(newBytes)
}

func fixOutOfBounds(text string, pageWidth float64, pageHeight float64, pageWidthCheck float64, ret *thehive.HiveCaseResp, tlp string, startPoint float64) {
	var err error
	desc := cleanupText(text)
	pdf.SetX(startPoint)

	realWidth, _ := pdf.MeasureTextWidth(desc)
	if realWidth > pageWidth {
		for _, item := range strings.Split(desc, " ") {
			itemWidth, _ := pdf.MeasureTextWidth(item)

			if pdf.GetX()+itemWidth > pageWidth+pageWidthCheck-400 {
				pdf.Br(20)
				pdf.SetX(startPoint)
			}

			if pdf.GetY() > pageHeight-200 {
				err = pdf.SetFont("OpenSans", "", 10)
				if err != nil {
					log.Print(err.Error())
					return
				}

				addLocalPage(ret, tlp, pageHeight)

			}
			pdf.Cell(nil, fmt.Sprintf(" %s", item))
		}
		pdf.Br(40)

	} else {
		pdf.Cell(nil, desc)
		pdf.Br(40)
	}

}

func addFonts() error {
	// Add fonts to available stuff
	err := pdf.AddTTFFont("OpenSans", "ttf/OpenSans-Regular.ttf")
	if err != nil {
		log.Print(err.Error())
		return err
	}

	err = pdf.AddTTFFont("OpenSans-Italic", "ttf/OpenSans-Italic.ttf")
	if err != nil {
		log.Print(err.Error())
		return err
	}

	err = pdf.AddTTFFont("OpenSans-Bold", "ttf/OpenSans-Bold.ttf")
	if err != nil {
		log.Print(err.Error())
		return err
	}

	return nil
}

func addLocalPage(ret *thehive.HiveCaseResp, tlp string, pageHeight float64) {
	pdf.AddPage()

	err := pdf.SetFont("OpenSans", "", 10)
	if err != nil {
		log.Print(err.Error())
		return
	}

	pdf.SetX(20)
	pdf.SetY(pageHeight - 20)
	pdf.SetTextColor(1, 1, 1)
	pdf.Cell(nil, fmt.Sprintf("TLP: %s", tlp))
	pdf.SetX(0)
	pdf.SetY(0)
	pdf.SetTextColor(100, 100, 100)
	pdf.Image("images/thehive-logo.png", 20, 20, nil) //print image
	// SetX = totWidth-length(title)-50 > 200~

	pdf.SetX(200) //move current location
	pdf.SetY(20)  //move current location

	titleWidthFloat, _ := pdf.MeasureTextWidth(ret.Title)
	titleWidthInt := int(titleWidthFloat)
	widthCheck := 80

	if titleWidthInt < widthCheck {
		pdf.Cell(nil, ret.Title)
	} else {
		for cnt, item := range []byte(ret.Title) {
			if cnt%widthCheck == 0 {
				pdf.Br(10)
				pdf.SetX(200) //move current location
			}

			pdf.Cell(nil, string(item))
		}
	}

	pdf.SetX(20)
	pdf.SetY(100)
}

func GeneratePdf(hive thehive.Hivedata, ret *thehive.HiveCaseResp) {
	var err error

	pdfName := fmt.Sprintf("reports/%d.pdf", ret.CaseId)
	log.Printf("Starting generation of report %s", pdfName)
	pageWidth := 595.28
	pageHeight := 841.89
	pageWidthCheck := 2000.0

	// Create object
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{W: pageWidth, H: pageHeight},
	})

	// Adds all fonts
	err = addFonts()
	if err != nil {
		log.Print(err.Error())
		return
	}

	err = pdf.SetFont("OpenSans", "", 10)
	if err != nil {
		log.Print(err.Error())
		return
	}

	tlp := ""
	if ret.Tlp == 1 {
		tlp = "White"
	} else if ret.Tlp == 2 {
		tlp = "Green"
	} else if ret.Tlp == 3 {
		tlp = "Amber"
	} else if ret.Tlp == 4 {
		tlp = "Red"
	}

	addLocalPage(ret, tlp, pageHeight)
	//pdf.SetX(20) //move current location
	//pdf.SetY(50)

	err = pdf.SetFont("OpenSans", "", 25)
	if err != nil {
		log.Print(err.Error())
		return
	}

	pdf.SetTextColor(0, 92, 197)
	//pdf.Br(55)
	pdf.SetX(20)
	desc := ret.Title
	realWidth, _ := pdf.MeasureTextWidth(ret.Title)
	if realWidth > pageWidth {
		// FIXME - problem here with long words lol
		for _, item := range strings.Split(desc, " ") {
			itemWidth, _ := pdf.MeasureTextWidth(item)

			if pdf.GetX()+itemWidth > pageWidth+pageWidthCheck-400 {
				pdf.Br(30)
				pdf.SetX(20)
			}

			// Handles newlines and shit too
			if pdf.GetY() > pageHeight-200 {
				err = pdf.SetFont("OpenSans", "", 10)
				if err != nil {
					log.Print(err.Error())
					return
				}
				addLocalPage(ret, tlp, pageHeight)
			}
			pdf.Cell(nil, fmt.Sprintf(" %s", item))
		}
		pdf.Br(40)

	} else {
		pdf.Cell(nil, desc)
		pdf.Br(40)
	}
	err = pdf.SetFont("OpenSans-Italic", "", 14)
	if err != nil {
		log.Print(err.Error())
		return
	}

	// Exact position

	// Clean up description

	// Description formatting
	// Basic info
	pdf.SetTextColor(1, 1, 1)
	pdf.SetX(20)
	pdf.Cell(nil, fmt.Sprintf("Owner: %s", ret.Owner))
	pdf.Br(20)
	pdf.SetX(20)
	pdf.Cell(nil, fmt.Sprintf("Status: %s", ret.Status))
	pdf.Br(20)
	pdf.SetX(20)
	pdf.Cell(nil, fmt.Sprintf("Case Number: %d", ret.CaseId))
	pdf.Br(20)
	pdf.SetX(20)
	pdf.Cell(nil, fmt.Sprintf("Id: %s", ret.Id))
	pdf.Br(20)
	pdf.SetX(20)
	pdf.Cell(nil, fmt.Sprintf("Tags: %s", strings.Join(ret.Tags, ", ")))
	pdf.Br(20)
	pdf.SetX(20)

	//curtime := strconv.FormatInt(time.Now().Unix(), 10)
	//newcurtime, _ := strconv.ParseInt(fmt.Sprintf("%s000", curtime), 10, 64)
	// 1565613531463

	pdf.Cell(nil, fmt.Sprintf("Severity: "))
	severity := ""
	if ret.Severity == 1 {
		severity = "Low"
		pdf.SetTextColor(0, 255, 0)
	} else if ret.Severity == 2 {
		severity = "Medium"
		pdf.SetTextColor(255, 255, 0)
	} else if ret.Severity == 3 {
		severity = "High"
		pdf.SetTextColor(255, 0, 0)
	} else if ret.Severity == 4 {
		severity = "Critical"
		pdf.SetTextColor(1, 1, 1)
	}
	pdf.Cell(nil, severity)
	pdf.Br(20)
	pdf.SetX(20)

	pdf.SetTextColor(1, 1, 1)
	pdf.Cell(nil, fmt.Sprintf("TLP: "))

	if ret.Tlp == 1 {
		pdf.SetTextColor(1, 1, 1)
	} else if ret.Tlp == 2 {
		pdf.SetTextColor(0, 255, 0)
	} else if ret.Tlp == 3 {
		pdf.SetTextColor(255, 255, 0)
	} else if ret.Tlp == 4 {
		pdf.SetTextColor(255, 0, 0)
	}
	pdf.Cell(nil, tlp)
	pdf.Br(20)
	pdf.SetX(20)
	pdf.SetTextColor(1, 1, 1)

	timestamp := fmt.Sprintf("%d", ret.CreatedAt)
	if timestamp != "" {
		timestamp := timestamp[0 : len(timestamp)-3]
		i, err := strconv.ParseInt(timestamp, 10, 64)
		if err == nil {
			tm := time.Unix(i, 0)
			pdf.Cell(nil, fmt.Sprintf("Creation date: %s", tm))
			pdf.Br(20)
			pdf.SetX(20)
		}

	}

	pdf.SetTextColor(1, 1, 1)
	pdf.SetX(20)
	// Custom fields

	//pdf.Cell(nil, fmt.Sprintf("Description: %s", ret.Description))
	//pdf.Br(20)
	//pdf.SetX(20)

	pdf.SetTextColor(1, 1, 1)
	if ret.Status == "Resolved" {
		err = pdf.SetFont("OpenSans-Italic", "", 16)
		if err != nil {
			log.Print(err.Error())
			return
		}

		pdf.Line(20, pdf.GetY(), 575, pdf.GetY())
		pdf.Br(20)
		pdf.SetX(20)
		pdf.Cell(nil, fmt.Sprintf("Close code: %s", ret.ResolutionStatus))
		pdf.Br(20)
		pdf.SetX(20)
		pdf.Cell(nil, fmt.Sprintf("Close reason: %s", ret.Summary))
		pdf.Br(20)
		pdf.SetX(20)

		timestamp = fmt.Sprintf("%d", ret.EndDate)
		if timestamp != "" {
			timestamp := timestamp[0 : len(timestamp)-3]
			i, err := strconv.ParseInt(timestamp, 10, 64)
			if err == nil {
				tm := time.Unix(i, 0)
				pdf.Cell(nil, fmt.Sprintf("Close date: %s", tm))
				pdf.Br(20)
				pdf.SetX(20)
			}

		}

		if ret.ResolutionStatus == "TruePositive" {
			if ret.ImpactStatus == "WithImpact" {
				pdf.Cell(nil, fmt.Sprintf("Impact: True"))
				pdf.Br(20)
				pdf.SetX(20)
			}
		}
	}

	pdf.SetTextColor(100, 100, 100)
	err = pdf.SetFont("OpenSans-Italic", "", 14)
	if err != nil {
		log.Print(err.Error())
		return
	}
	pdf.Br(20)
	pdf.SetX(20)
	pdf.Line(20, pdf.GetY()-20, 575, pdf.GetY()-20)
	fixOutOfBounds(ret.Description, pageWidth, pageHeight, pageWidthCheck, ret, tlp, 20.0)

	// Handle observables page
	artifacts, err := hive.GetCaseArtifacts(ret.Id)
	if len(artifacts.Detail) > 0 {
		addLocalPage(ret, tlp, pageHeight)
		err = pdf.SetFont("OpenSans", "", 25)
		if err != nil {
			log.Print(err.Error())
			return
		}
		pdf.SetTextColor(0, 92, 197)

		//pdf.Br(50)
		//pdf.SetX(20)
		pdf.Cell(nil, "Observables")
		pdf.Br(35)
		pdf.SetX(20)

		err = pdf.SetFont("OpenSans", "", 12)
		if err != nil {
			log.Print(err.Error())
			return
		}

		pdf.SetTextColor(1, 1, 1)
		pdf.Cell(nil, fmt.Sprintf("Type, data, sighted, ioc"))
		pdf.Br(20)
		pdf.SetX(20)

		// Find the different types
		types := []string{}
		for _, artifact := range artifacts.Detail {
			typeCheck := true
			for _, curtype := range types {
				if curtype == artifact.DataType {
					typeCheck = false
				}

			}
			if typeCheck {
				types = append(types, artifact.DataType)
			}
		}

		// Slow typebased sorting
		for _, curtype := range types {
			for _, artifact := range artifacts.Detail {
				if artifact.DataType != curtype {
					continue
				}

				pdf.Cell(nil, fmt.Sprintf("%s %s %t %t", artifact.DataType, artifact.Data, artifact.Sighted, artifact.Ioc))
				pdf.Br(20)
				pdf.SetX(20)

			}
		}
	}

	// Tasklogs
	tasks, err := hive.GetCaseTasks(ret.Id)
	if len(tasks.Detail) > 0 {
		// NEW PAGE
		err = pdf.SetFont("OpenSans", "", 10)
		if err != nil {
			log.Print(err.Error())
			return
		}

		addLocalPage(ret, tlp, pageHeight)

		err = pdf.SetFont("OpenSans", "", 25)
		if err != nil {
			log.Print(err.Error())
			return
		}
		pdf.SetTextColor(0, 92, 197)

		//pdf.Br(50)
		//pdf.SetX(20)
		pdf.Cell(nil, fmt.Sprintf("Tasklogs (%d)", len(tasks.Detail)))
		pdf.Br(20)
		pdf.SetX(20)

		pdf.SetTextColor(1, 1, 1)
		for count, item := range tasks.Detail {
			tasklogs, err := hive.GetTaskLogs(item.Id)
			if err != nil {
				log.Print("Tasklog error: %s", err)
				return
			}

			// Always new page when tasklog with data
			if len(tasklogs.Detail) > 0 && count != 0 {
				addLocalPage(ret, tlp, pageHeight)
			}

			if count == 0 {
				pdf.Line(20, pdf.GetY()+20, 575, pdf.GetY()+20)
				pdf.Br(25)
				pdf.SetX(20)
			}

			// Get tasks
			err = pdf.SetFont("OpenSans-Bold", "", 14)
			if err != nil {
				log.Print(err.Error())
				return
			}

			ownerNameLength := pdf.GetX()
			if pdf.GetY()+20 > pageHeight-200 {
				ownerNameLength := pdf.GetX()
				err = pdf.SetFont("OpenSans", "", 10)
				if err != nil {
					log.Print(err.Error())
					return
				}

				addLocalPage(ret, tlp, pageHeight)

				//pdf.Br(50)
				pdf.SetX(ownerNameLength)
				pdf.SetTextColor(100, 100, 100)
				err = pdf.SetFont("OpenSans-Italic", "", 14)
				if err != nil {
					log.Print(err.Error())
					return
				}
			}
			pdf.SetTextColor(1, 1, 1)
			pdf.Cell(nil, fmt.Sprintf("Title: %s", item.Title))
			pdf.Br(20)
			pdf.SetX(20)
			pdf.Cell(nil, fmt.Sprintf("Owner: %s", item.Owner))
			pdf.Br(20)
			pdf.SetX(20)
			pdf.Cell(nil, fmt.Sprintf("Status: %s", item.Status))
			pdf.Br(20)
			pdf.SetX(20)
			pdf.Cell(nil, fmt.Sprintf("Type: %s", item.Type))
			pdf.Br(20)
			pdf.SetX(20)

			pdf.Cell(nil, fmt.Sprintf("Description: "))
			pdf.Br(20)
			pdf.SetX(20)
			pdf.SetTextColor(100, 100, 100)
			fixOutOfBounds(item.Description, pageWidth, pageHeight, pageWidthCheck, ret, tlp, 20.0)
			pdf.SetTextColor(1, 1, 1)

			err = pdf.SetFont("OpenSans-Italic", "", 14)
			if err != nil {
				log.Print("FontsettingFontsetting  error: %s", err)
				return
			}

			if len(tasklogs.Detail) <= 0 {
				if pdf.GetY()+20 > pageHeight-200 {
					err = pdf.SetFont("OpenSans", "", 10)
					if err != nil {
						log.Print(err.Error())
						return
					}

					addLocalPage(ret, tlp, pageHeight)

					pdf.SetX(ownerNameLength)
					pdf.SetTextColor(100, 100, 100)
					err = pdf.SetFont("OpenSans-Italic", "", 14)
					if err != nil {
						log.Print(err.Error())
						return
					}
				}
				pdf.SetLineWidth(1)
				pdf.Line(20, pdf.GetY()+20, 575, pdf.GetY()+20)
				pdf.Br(25)
				pdf.SetX(20)
				pdf.Cell(nil, "This tasklog has no comments.")
				pdf.SetLineWidth(1)
				pdf.Line(20, pdf.GetY()+20, 575, pdf.GetY()+20)
				pdf.Br(25)
				pdf.SetX(20)
			} else {
				for _, tasklog := range tasklogs.Detail {
					pdf.SetLineWidth(1)
					pdf.Line(20, pdf.GetY()+20, 575, pdf.GetY()+20)
					pdf.Br(25)
					pdf.SetX(20)
					err = pdf.SetFont("OpenSans-Bold", "", 14)
					if err != nil {
						log.Print(err.Error())
						return
					}

					pdf.SetTextColor(1, 1, 1)
					pdf.Cell(nil, fmt.Sprintf("%s: ", item.Owner))
					ownerNameLength := pdf.GetX()

					pdf.SetTextColor(100, 100, 100)
					err = pdf.SetFont("OpenSans-Italic", "", 14)
					if err != nil {
						log.Print(err.Error())
						return
					}

					// Message, owner,
					//fixOutOfBounds(tasklog.Message, pageWidth, pageHeight, pageWidthCheck, ret, tlp, ownerNameLength)
					desc = cleanupText(tasklog.Message)

					realWidth, _ := pdf.MeasureTextWidth(desc)
					if realWidth > pageWidth {
						for _, item := range strings.Split(desc, " ") {
							itemWidth, _ := pdf.MeasureTextWidth(item)

							// Arbitrary number much
							if pdf.GetX()+itemWidth > pageWidth+pageWidthCheck-400 {
								pdf.Br(20)
								pdf.SetX(ownerNameLength)
							}

							if pdf.GetY()+20 > pageHeight-100 {
								err = pdf.SetFont("OpenSans", "", 10)
								if err != nil {
									log.Print(err.Error())
									return
								}

								addLocalPage(ret, tlp, pageHeight)

								//pdf.Br(50)
								pdf.SetX(ownerNameLength)
								pdf.SetTextColor(100, 100, 100)
								err = pdf.SetFont("OpenSans-Italic", "", 14)
								if err != nil {
									log.Print(err.Error())
									return
								}
							}

							pdf.Cell(nil, fmt.Sprintf(" %s", item))
						}

						pdf.Br(40)
						pdf.SetX(ownerNameLength)

					} else {
						pdf.Cell(nil, desc)
						pdf.Br(40)
						pdf.SetX(ownerNameLength)
					}
				}

				pdf.Br(20)
				pdf.SetX(20)
			}
		}
	}

	// FIXME - pages

	// IOCs
	//hive.GetCaseArtifacts(ret.Id)

	// Exact position

	err = pdf.WritePdf(pdfName)
	if err != nil {
		//log.Println(err)
		err = os.Mkdir("reports", os.ModePerm)
		if err != nil {
			log.Printf("Failed generating: %s", err)
			time.Sleep(30 * time.Second)
			os.Exit(3)
		}

		err = pdf.WritePdf(pdfName)
		if err != nil {
			log.Printf("Failed generating: %s", err)
			time.Sleep(30 * time.Second)
			os.Exit(3)
		} else {
			log.Printf("GENERATED %s!\n", pdfName)
			time.Sleep(10 * time.Second)
		}
	} else {
		log.Printf("GENERATED %s!\n", pdfName)
		time.Sleep(10 * time.Second)
	}
}

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

		// FIXME
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
			//log.Printf(string(ret.Raw))
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
