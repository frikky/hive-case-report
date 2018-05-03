package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/aodin/date"
	"github.com/bitly/go-simplejson"
	"github.com/frikky/hive4go"
	//"github.com/jinzhu/now"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func getThisSearch(fromTime string, toTime string) (string, error) {
	dateFrom, err := date.Parse(fromTime)
	if err != nil {
		return "", err
	}
	dateTo, err := date.Parse(toTime)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"query": {"_and":[{"_string": "startDate: [ %s TO %s ]"}]}}`, fmt.Sprintf("%d000", dateFrom.Unix()), fmt.Sprintf("%d000", dateTo.Unix())), nil
}

func getreport(hive thehive.Hivedata, from string, to string) error {
	// Check if early in month, get previousmonth
	// Check if late in month, get this month
	var tmpjson string
	var err error

	tmpjson, err = getThisSearch(from, to)
	if err != nil {
		return err
	}

	jsonBlob := []byte(tmpjson)
	test, err := hive.FindCases(jsonBlob)
	if err != nil {
		return err
	}

	f, err := os.Create("cases.csv")
	if err != nil {
		return err
	}

	defer f.Close()
	csvWriter := csv.NewWriter(f)

	parseAllCases(test.Raw, csvWriter).Flush()

	return nil
}

func checkTag(tag string) bool {
	allTags := []string{
		"unauthorized application",
		"fraud",
		"phishing",
		"lost/stolen computer",
		"suspicious user activity",
		"operational issue",
		"malicious code",
		"reconnaissance",
		"policy violation",
		"vulnerability",
		"ddos",
	}

	tag = strings.ToLower(tag)
	for _, element := range allTags {
		if element == tag {
			return true
		}
	}
	return false
}

func convertSeverity(severity int) string {
	if severity == 1 {
		return "Low"
	} else if severity == 2 {
		return "Medium"
	} else if severity == 3 {
		return "High"
	}

	return "Low"
}

func parseAllCases(inputdata []byte, csvWriter *csv.Writer) *csv.Writer {
	jsondata, _ := simplejson.NewJson(inputdata)
	csvWriter.Write([]string{"Week", "ID", "Date Created", "Status", "Severity", "Title", "Tags"})
	var allTags = make(map[string]int)

	for _, newelement := range jsondata.MustArray() {
		element, _ := json.Marshal(newelement)
		newjsondata, _ := simplejson.NewJson(element)
		//createTime := int64(newjsondata.Get("createdAt").MustInt())

		caseId := strconv.Itoa(newjsondata.Get("caseId").MustInt())
		status := newjsondata.Get("status").MustString()

		createDate := newjsondata.Get("createdAt").MustInt()
		fromStr := fmt.Sprintf("%d", createDate)
		newFrom, _ := strconv.Atoi(fromStr[:len(fromStr)-3])
		createTime := time.Unix(int64(newFrom), 0)

		_, week := createTime.ISOWeek()

		/*
			// Handles close dates for closed stuff
			closeDate := "0"
			if status != "Open" {
				closeDate = strconv.Itoa(newjsondata.Get("endDate").MustInt())
			}
		*/

		tagArray := newjsondata.Get("tags").MustArray()

		for _, element := range tagArray {
			allTags[fmt.Sprintf("%v", element)] += 1
		}

		tagInterface := fmt.Sprintf("%v", tagArray)
		tags := strings.Join(strings.Split(tagInterface[1:len(tagInterface)-1], " "), ",")

		//severity := strconv.Itoa(
		severity := convertSeverity(newjsondata.Get("severity").MustInt())
		title := newjsondata.Get("title").MustString()

		//location, _ := time.LoadLocation("Europe/Oslo")

		csvWriter.Write([]string{
			strconv.Itoa(week),
			caseId,
			//createTime.In(location).String(),
			createTime.String(),
			status,
			severity,
			title,
			tags,
		})
	}

	csvWriter.Write([]string{})
	csvWriter.Write([]string{"TagName", "Amount"})

	// Below sorts and prints tags
	sortedList := map[int][]string{}
	var a []int

	for key, value := range allTags {
		sortedList[value] = append(sortedList[value], key)
	}

	for key := range sortedList {
		a = append(a, key)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(a)))
	for _, k := range a {
		for _, s := range sortedList[k] {
			if checkTag(s) {
				csvWriter.Write([]string{s, strconv.Itoa(k)})
			}
		}
	}

	return csvWriter
}
