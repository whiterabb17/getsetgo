# Filesearch utility
sfs is short for "simple file search"  

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
sfs somefile
```

3. Output will be
```
/Users/username/somedir/somefile.txt
/Users/username/somedir/somedir2/somefile.py
/Users/username/somedir/somedir2/somefile
...
```