# Go SSH

## run cmd
```
sh := NewSSHExecutor(username, password, ipaddr, port)
if err := sh.Cmd(cmd); err != nil {
    fmt.Println(err)
}
```

## run cmd with return
```
sh := NewSSHExecutor(username, password, ipaddr, port)
rtn, err := sh.CmdGet(cmd)
fmt.Print(rtn, err)
}
```

## Put
```
file1 := FilePut{LocalFile: "/path/to/file1", RemoteDir: "/path/to/dir"}
file2 := FilePut{LocalFile: "/path/to/file2", RemoteDir: "/path/to/dir"}
sf := NewSFTPExecutor(username, password, ipaddr, port)
if err := sf.Put([]FilePut{file1, file2}); err != nil {
    fmt.Println(err)
}
```

## Get
```
file1 := FileGet{LocalDir: "/path/to/dir", RemoteFile: "/path/to/file1"}
file2 := FileGet{LocalDir: "/path/to/dir", RemoteFile: "/path/to/file2"}
sf := NewSFTPExecutor(username, password, ipaddr, port)
if err := sf.Get([]FileGet{file1, file2}); err != nil {
    fmt.Println(err)
}
```

## Interactive Shell
```
shell := NewInteractiveShell(username, password, ipaddr, port)
_ = shell.InvokeShell(120, 120)
defer shell.ChanClose()
ch := make(chan string)
go shell.ChanRcv(ch)
go func(ch chan string) {
	for s := range ch {
		fmt.Print(s)
	}
}(ch)
time.Sleep(100 * time.Millisecond)
shell.ChanSend("ls /")
shell.ChanSend("exit")
time.Sleep(10 * time.Second)
```