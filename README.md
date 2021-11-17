# Go SSH

## run cmd
```
host := NewHost(username, password, ipaddr, port)
if err := host.Cmd(cmd); err != nil {
    fmt.Println(err)
}
```

## run cmd with return
```
host := NewHost(username, password, ipaddr, port)
rtn, err := host.CmdGet(cmd)
fmt.Print(rtn, err)
}
```

## Put
```
file1 := FilePut{LocalFile: "/path/to/file1", RemoteDir: "/path/to/dir"}
file2 := FilePut{LocalFile: "/path/to/file2", RemoteDir: "/path/to/dir"}
host := NewHost(username, password, ipaddr, port)
if err := host.Put([]FilePut{file1, file2}); err != nil {
    fmt.Println(err)
}
```

## Get
```
file1 := FileGet{LocalDir: "/path/to/dir", RemoteFile: "/path/to/file1"}
file2 := FileGet{LocalDir: "/path/to/dir", RemoteFile: "/path/to/file2"}
host := NewHost(username, password, ipaddr, port)
if err := host.Get([]FileGet{file1, file2}); err != nil {
    fmt.Println(err)
}
```