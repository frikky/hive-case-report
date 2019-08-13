# PDF Report generator for thehive cases
This is a (bad) generator for TheHive cases, which generates a PDF of the case. Its based on just traversing all data in a given case, including observables and task logs etc.

# Usage - linux 
1. Add your API key and TheHive URL to config.json 
2. 
```bash 
go run report.go
```
![caseId1](https://github.com/frikky/hive-case-report/blob/master/images/reporting.PNG?raw=true)

3. Open e.g. ./reports/513.pdf 

## Sample
![firstpage](https://github.com/frikky/hive-case-report/blob/master/images/Firstpage.PNG?raw=true)
![morepages](https://github.com/frikky/hive-case-report/blob/master/images/Morepages.PNG?raw=true)

## Usage - windows export
```bash
# Generate report.exe for x86_64 
sh compile.sh
```

### Missing: 
* Proper newline handling
* Custom fields
* Artifact parsing and sorting (they're included without)
* Metrixx
