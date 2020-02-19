# disableExpiredDeviceRuleTool
##disableExpiredDeviceRuleTool##
- add the cronjob you want to run to mysql table: crontab
```
ex:
# cronjobId, cronjobName, cronCmd, freq
'2', 'CheckLicense', 'cd /home/zyxel/vicky/zyxelProjects/docker-test/disableExpiredDeviceRuleTool/ && ./disableTool 127.0.0.1:3308 root root http://127.0.0.1:8080 2000-01-01 2019-11-01', '*/1 * * * *'

```
- crontab log files' format will be like {cronjobName}_log_2020_02_19/2020-02-19_102319.log
- build Dockerfile, and run it.
- Dockerfile will build all go files cronjobs need, check checkCronList.go to know how they run.
- After you see "crontab added!", you can use "crontab -e" and "grep CRON /var/log/syslog" to check if cronjobs added successfully. 
