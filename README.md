[![Build Status](https://travis-ci.org/mailgun/glogutils.png)](https://travis-ci.org/mailgun/glogutils)

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

Example
-------

```go
// This function starts cleaning up after glog library, periodically removing logs
// that are no longer used
func startLogsCleanup(period time.Duration) error {
	if glogutils.LogDir() != "" {
		glog.Infof("Starting log cleanup go routine with period: %s", period)
		go func() {
			t := time.Tick(period)
			for {
				select {
				case <-t:
					glog.Infof("Start cleaning up the logs")
					err := glogutils.CleanupLogs()
					if err != nil {
						glog.Errorf("Failed to clean up the logs: %s", err)
						return
					}
				}
			}
		}()
	}
	return nil
}
```

License
-------

Apache2
