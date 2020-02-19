# disableExpiredDeviceRuleTool
##disableExpiredDeviceRuleTool##
- add the cronjob you want to run to mysql table; crontab
- build Dockerfile, and run it.
- Dockerfile will build all go files cronjobs need, check checkCronList.go to know how they run.
