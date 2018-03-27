package main

import (
	"fmt"
	//wordwrap "github.com/mitchellh/go-wordwrap"

	"github.com/bitly/go-simplejson"
	"strings"
	//findfont "github.com/flopp/go-findfont"
	"github.com/frikky/hive4go"
	"github.com/signintech/gopdf"
	"io/ioutil"
	//"strings"
	"log"
	"os"
	"time"
)

func getTime() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

func getHiveLogin() (string, string) {
	var err error
	var configpath string

	configpath = "config.json"

	file, err := ioutil.ReadFile(configpath)
	if err != nil {
		//log.Fatal(err)
		fmt.Printf("%s Error getting hive: %s\n", getTime(), err)
	}

	jsondata, err := simplejson.NewJson(file)
	if err != nil {
		fmt.Printf("%s Error converting login to json: %s\n", getTime(), err)
	}

	return jsondata.Get("hiveurl").MustString(), jsondata.Get("hiveapikey").MustString()
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

func generatePdf(hive thehive.Hivedata, ret *thehive.HiveCaseResp) {
	//fmt.Println(string(ret.Raw))
	pageWidth := 595.28
	pageHeight := 841.89
	pageWidthCheck := 1600.0

	// Create object
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: gopdf.Rect{W: pageWidth, H: pageHeight},
	})

	// Add fonts to available stuff
	err := pdf.AddTTFFont("OpenSans", "ttf/OpenSans-Regular.ttf")
	if err != nil {
		log.Print(err.Error())
		return
	}

	err = pdf.AddTTFFont("OpenSans-Italic", "ttf/OpenSans-Italic.ttf")
	if err != nil {
		log.Print(err.Error())
		return
	}

	err = pdf.AddTTFFont("OpenSans-Bold", "ttf/OpenSans-Bold.ttf")
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

	pdf.AddPage()

	pdf.SetX(20)
	pdf.SetY(pageHeight - 20)
	pdf.SetTextColor(1, 1, 1)
	pdf.Cell(nil, fmt.Sprintf("TLP: %s", tlp))
	pdf.SetX(0)
	pdf.SetY(0)
	pdf.SetTextColor(100, 100, 100)
	pdf.Image("thehive-logo.png", 20, 20, nil) //print image
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

	//pdf.SetX(20) //move current location
	//pdf.SetY(50)

	err = pdf.SetFont("OpenSans", "", 25)
	if err != nil {
		log.Print(err.Error())
		return
	}

	pdf.SetTextColor(0, 92, 197)
	pdf.Br(55)
	pdf.SetX(20)
	desc := ret.Title
	realWidth, _ := pdf.MeasureTextWidth(ret.Title)
	if realWidth > pageWidth {
		for _, item := range strings.Split(desc, " ") {
			itemWidth, _ := pdf.MeasureTextWidth(item)

			//fmt.Println(pdf.GetX()+itemWidth, pageWidth)
			if pdf.GetX()+itemWidth > pageWidth+pageWidthCheck {
				pdf.Br(30)
				pdf.SetX(20)
			}

			// Handles newlines and shit too
			if pdf.GetY() > pageHeight-50 {
				err = pdf.SetFont("OpenSans", "", 10)
				if err != nil {
					log.Print(err.Error())
					return
				}
				pdf.AddPage()

				pdf.SetX(20)
				pdf.SetY(pageHeight - 20)
				pdf.SetTextColor(1, 1, 1)
				pdf.Cell(nil, fmt.Sprintf("TLP: %s", tlp))
				pdf.SetX(0)
				pdf.SetY(0)
				pdf.SetTextColor(100, 100, 100)
				pdf.Image("thehive-logo.png", 20, 20, nil) //print image
				// SetX = totWidth-length(title)-50 > 200~

				pdf.SetX(200) //move current location
				pdf.SetY(20)  //move current location

				titleWidthFloat, _ := pdf.MeasureTextWidth(ret.Title)
				titleWidthInt := int(titleWidthFloat)
				widthCheck := 70

				if titleWidthInt < widthCheck {
					fmt.Println("LESS")
					pdf.Cell(nil, ret.Title)
				} else {
					fmt.Println("More")

					for cnt, item := range []byte(ret.Title) {
						fmt.Println(cnt, widthCheck)
						if cnt%widthCheck == 0 {
							pdf.Br(10)
							pdf.SetX(200) //move current location
						}

						pdf.Cell(nil, string(item))
					}
				}
			}
			pdf.Cell(nil, fmt.Sprintf(" %s", item))
		}
		pdf.Br(40)

	} else {
		pdf.Cell(nil, desc)
		pdf.Br(40)
	}

	// Exact position
	pdf.SetTextColor(100, 100, 100)

	err = pdf.SetFont("OpenSans-Italic", "", 14)
	if err != nil {
		log.Print(err.Error())
		return
	}

	// Clean up description
	//fmt.Println(string(ret.Raw))
	desc = cleanupText(ret.Description)
	pdf.SetX(20)

	realWidth, _ = pdf.MeasureTextWidth(desc)
	if realWidth > pageWidth {
		for _, item := range strings.Split(desc, " ") {
			itemWidth, _ := pdf.MeasureTextWidth(item)

			//fmt.Println(pdf.GetX()+itemWidth, pageWidth)
			if pdf.GetX()+itemWidth > pageWidth+pageWidthCheck {
				pdf.Br(20)
				pdf.SetX(20)
			}

			if pdf.GetY() > pageHeight-50 {
				err = pdf.SetFont("OpenSans", "", 10)
				if err != nil {
					log.Print(err.Error())
					return
				}
				pdf.AddPage()

				pdf.SetX(20)
				pdf.SetY(pageHeight - 20)
				pdf.SetTextColor(1, 1, 1)
				pdf.Cell(nil, fmt.Sprintf("TLP: %s", tlp))
				pdf.SetX(0)
				pdf.SetY(0)
				pdf.SetTextColor(100, 100, 100)
				pdf.Image("thehive-logo.png", 20, 20, nil) //print image
				// SetX = totWidth-length(title)-50 > 200~

				pdf.SetX(200) //move current location
				pdf.SetY(20)  //move current location

				titleWidthFloat, _ := pdf.MeasureTextWidth(ret.Title)
				titleWidthInt := int(titleWidthFloat)
				widthCheck := 70

				if titleWidthInt < widthCheck {
					fmt.Println("LESS")
					pdf.Cell(nil, ret.Title)
				} else {
					fmt.Println("More")

					for cnt, item := range []byte(ret.Title) {
						fmt.Println(cnt, widthCheck)
						if cnt%widthCheck == 0 {
							pdf.Br(10)
							pdf.SetX(200) //move current location
						}

						pdf.Cell(nil, string(item))
					}
				}
			}
			pdf.Cell(nil, fmt.Sprintf(" %s", item))
		}
		pdf.Br(40)

	} else {
		pdf.Cell(nil, desc)
		pdf.Br(40)
	}

	pdf.Line(20, pdf.GetY()-20, 575, pdf.GetY()-20)

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
	pdf.Cell(nil, fmt.Sprintf("Id: %s", ret.Id))
	pdf.Br(20)
	pdf.SetX(20)
	pdf.Cell(nil, fmt.Sprintf("Tags: %s", strings.Join(ret.Tags, ", ")))
	pdf.Br(20)
	pdf.SetX(20)

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

	// Tasklogs
	tasks, err := hive.GetCaseTasks(ret.Id)

	if len(tasks.Detail) > 0 {
		// NEW PAGE
		err = pdf.SetFont("OpenSans", "", 10)
		if err != nil {
			log.Print(err.Error())
			return
		}
		pdf.AddPage()

		pdf.SetX(20)
		pdf.SetY(pageHeight - 20)
		pdf.SetTextColor(1, 1, 1)
		pdf.Cell(nil, fmt.Sprintf("TLP: %s", tlp))
		pdf.SetX(0)
		pdf.SetY(0)
		pdf.SetTextColor(100, 100, 100)
		pdf.Image("thehive-logo.png", 20, 20, nil) //print image
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

		err = pdf.SetFont("OpenSans", "", 25)
		if err != nil {
			log.Print(err.Error())
			return
		}
		pdf.SetTextColor(0, 92, 197)

		pdf.Br(50)
		pdf.SetX(20)
		pdf.Cell(nil, "Tasklogs")
		pdf.Br(20)
		pdf.SetX(20)

		pdf.SetTextColor(1, 1, 1)
		for _, item := range tasks.Detail {
			// Get tasks
			err = pdf.SetFont("OpenSans-Bold", "", 14)
			if err != nil {
				log.Print(err.Error())
				return
			}

			pdf.Line(20, pdf.GetY()+20, 575, pdf.GetY()+20)
			pdf.Br(25)
			pdf.SetX(20)
			pdf.Cell(nil, fmt.Sprintf("Title: %s", item.Title))
			pdf.Br(20)
			pdf.SetX(20)
			pdf.Cell(nil, fmt.Sprintf("Owner: %s", item.Owner))
			pdf.Br(20)
			pdf.SetX(20)
			pdf.Cell(nil, fmt.Sprintf("Status: %s", item.Status))

			tasklogs, _ := hive.GetTaskLogs(item.Id)

			err = pdf.SetFont("OpenSans-Italic", "", 14)
			if err != nil {
				log.Print(err.Error())
				return
			}

			if len(tasklogs.Detail) <= 0 {
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
					pdf.Cell(nil, fmt.Sprintf("%s: ", item.Owner))
					ownerNameLength := pdf.GetX()

					pdf.SetTextColor(100, 100, 100)
					err = pdf.SetFont("OpenSans-Italic", "", 14)
					if err != nil {
						log.Print(err.Error())
						return
					}

					// Message, owner,
					desc = cleanupText(tasklog.Message)

					realWidth, _ := pdf.MeasureTextWidth(desc)
					if realWidth > pageWidth {
						for _, item := range strings.Split(desc, " ") {
							itemWidth, _ := pdf.MeasureTextWidth(item)

							// Arbitrary number much
							if pdf.GetX()+itemWidth > pageWidth+pageWidthCheck {
								pdf.Br(20)
								pdf.SetX(ownerNameLength)
							}

							if pdf.GetY()+20 > pageHeight-50 {
								err = pdf.SetFont("OpenSans", "", 10)
								if err != nil {
									log.Print(err.Error())
									return
								}

								pdf.AddPage()

								pdf.SetX(20)
								pdf.SetY(pageHeight - 20)
								pdf.SetTextColor(1, 1, 1)
								pdf.Cell(nil, fmt.Sprintf("TLP: %s", tlp))
								pdf.SetX(0)
								pdf.SetY(0)

								pdf.SetTextColor(100, 100, 100)
								pdf.Image("thehive-logo.png", 20, 20, nil) //print image
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
								pdf.Br(50)
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
	//fmt.Println(pdf.GetPages())
	// FIXME - IOCs etc

	// IOCs
	//hive.GetCaseArtifacts(ret.Id)

	// Exact position
	pdfName := fmt.Sprintf("%s.pdf", ret.Id)

	pdf.WritePdf(pdfName)
	fmt.Printf("GENERATED %s!\n", pdfName)
}

func main() {
	hiveurl, apikey := getHiveLogin()
	hive := thehive.CreateLogin(hiveurl, apikey, false)

	if len(os.Args) < 2 {
		fmt.Printf("Missing case.\nUsage: go run report.go CaseID\n")
		os.Exit(3)
	}
	fmt.Println(os.Args[1])

	ret, err := hive.GetCase(os.Args[1])
	//ret, err := hive.GetCase("AWJngikPX_yl8AikPKuN")
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}

	generatePdf(hive, ret)
}
