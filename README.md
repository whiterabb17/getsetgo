[![CodeFactor](https://www.codefactor.io/repository/github/whiterabb17/getsetgo/badge)](https://www.codefactor.io/repository/github/whiterabb17/getsetgo)

# Filesearch utility
<b>getsetgo</b> [ As in. On your marks, get set, go! ]<br>
Extremely fast file searching utility with the ability to copy found files to an archive.<br>
Time to search 912GB of a 1TB drive: 4.34.778min

### Dependencies
1. Golang installed
2. bin directory is present in gopath

### Deploy and use

1. Build and install
```
make install
```

2. Use from any directory. Current directory will be used as ROOT of search
```
getsetgo <path> <filename/extention>                          | Print output of found files
getsetgo <path> <filename/extention> output.txt               | Save output to a file
getsetgo <path> <filename/extention> folder                   | Copies all found files to given folder then zips the folder exfil
```
