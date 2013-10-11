Rationale
---------
Google logging library is a wonderful lib:

https://github.com/golang/glog

However the logs that it generates are never deleted.

glogutils adds one function to cleanup logs:

```go
glogutils.CleanupLogs() //Deletes files that are no longer active
```

and another one to understand if google logging logs to dir:

```go
glogutils.LogDir() // returns "" if log dir was not specified
```

License
-------

Apache2
